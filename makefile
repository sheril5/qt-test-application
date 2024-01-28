export GO111MODULE = on
APP_NAME?=qt-test-application

TOOLS_DIR		:= .tools/
GOLANGCI_LINT	:= ${TOOLS_DIR}github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.1${BIN_EXE}
GOTESTSUM		:= ${TOOLS_DIR}gotest.tools/gotestsum@v1.6.2${BIN_EXE}
OAPI_CODEGEN	:= ${TOOLS_DIR}github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.3.6${BIN_EXE}
GOIMPORTS		:= ${TOOLS_DIR}mvdan.cc/gofumpt/gofumports@v0.1.1${BIN_EXE}
BUILD_VERSION?=0.0.0-snapshot
.PHONY: test run build build-with-docker docker-build docker-push lint install-build-deps

${GOLANGCI_LINT} ${GOTESTSUM} ${OAPI_CODEGEN} ${GOIMPORTS}:
	$(eval TOOL=$(@:%${BIN_EXE}=%))
	@echo Installing ${TOOL}...
	@cd; GO111MODULE=on go install $(TOOL:${TOOLS_DIR}%=%)
	@mkdir -p $(dir ${TOOL})
	@cp ${GOBIN}/$(firstword $(subst @, ,$(notdir ${TOOL}))) ${TOOL}

GOPATH := $(shell go env GOPATH)

PACKAGES = $(shell go list ./... | grep -v /vendor/)


install-build-deps:
	go install -v $(LINTERS)

LINT_ONLY_NEW := --new-from-rev=HEAD~
LINT_FORMAT   := --out-format=checkstyle
LINT_OUTPUT   := > report/lint-report.xml
lint: ${GOLANGCI_LINT}				## Lint single service or package
	@echo Running lintter for mailserver
	@mkdir -p report
	${GOLANGCI_LINT} run -c=.golangci.yml ${LINT_FORMAT} ${LINT_ONLY_NEW} --build-tags integration,contract_test_consumer,contract_test_provider ./... ${LINT_OUTPUT}

UNIT_COVERAGE_OUTPUT   := > report/coverage.txt
test: lint
	mkdir -p builds
	env GO111MODULE=on go test -mod=vendor -race -coverprofile=${UNIT_COVERAGE_OUTPUT} ./...

cleanup:
	rm -r builds

run:
	env GOOS=linux \
	PORT=8088 \
	CGO_ENABLED=0 GO111MODULE=on /usr/local/go/bin/go run -mod=vendor  cmd/sample-http-server/main.go


vendor: go.mod go.sum
	@echo Downloading modules
	@go mod tidy
	@go mod download
	@go mod vendor
.PHONY: vendor

build: build_user  build_payment build_order
build_user:
	env GOOS=linux CGO_ENABLED=0 GO111MODULE=on /usr/local/go/bin/go build -mod=vendor -o builds/user cmd/user/main.go


build_payment:
	env GOOS=linux CGO_ENABLED=0 GO111MODULE=on /usr/local/go/bin/go build -mod=vendor -o builds/payment cmd/payment/main.go


build_order:
	env GOOS=linux CGO_ENABLED=0 GO111MODULE=on /usr/local/go/bin/go build -mod=vendor -o builds/order cmd/order/main.go


docker-build: build
	docker build --rm -t user -f ./docker/user/Dockerfile .
	docker build --rm -t order -f ./docker/order/Dockerfile .
	docker build --rm -t payment -f ./docker/payment/Dockerfile .

docker-build-images: docker-login  docker-build-images-user docker-build-images-order docker-build-images-payment

docker-login:
	docker login -u ${ARTIFACTORY_USER} -p ${ARTIFACTORY_PASSWORD}


docker-build-images-user:
	docker tag user snagarju/user:latest
	docker tag user snagarju/user:${BUILD_VERSION}
	docker push snagarju/user:latest
	docker push snagarju/user:${BUILD_VERSION}


docker-build-images-payment:
	docker tag payment snagarju/payment:latest
	docker tag payment snagarju/payment:${BUILD_VERSION}
	docker push snagarju/payment:latest
	docker push snagarju/payment:${BUILD_VERSION}


docker-build-images-order:
	docker tag order snagarju/order:latest
	docker tag order snagarju/order:${BUILD_VERSION}
	docker push snagarju/order:latest
	docker push snagarju/order:${BUILD_VERSION}


oapi-gen:	## Generate server code with oapi-codegen for single service
	@echo Generating server for mail server
	@mkdir -p api
	oapi-codegen -generate spec -package api ./openapi.yaml | gofumports > api/spec.go
	oapi-codegen -generate types -package api ./openapi.yaml | gofumports > api/models.go
	oapi-codegen -generate gin -package api ./openapi.yaml | gofumports > api/server.go

certs:
	mkdir -p hack
	openssl req  -new  -newkey rsa:2048  -nodes  -keyout ./hack/localhost.key  -out ./hack/localhost.csr
	openssl  x509  -req  -days 365  -in ./hack/localhost.csr  -signkey ./hack/localhost.key  -out ./hack/localhost.crt


recreate-cluster: delete-cluster create-cluster install-ingress

delete-cluster:
	kind delete cluster --name kind-local --kubeconfig ./test/kubeconfig.yaml

create-cluster:
	kind create  cluster --name kind-local --kubeconfig ./test/kubeconfig.yaml --config ./test/kind_config.yaml
	cp ./test/kubeconfig.yaml ~/kind/kubeconfig.yaml

delete-ingress:
	kubectl delete -f ./test/ingress.yaml --kubeconfig ./test/kubeconfig.yaml

install-ingress:
	kubectl apply -f ./test/ingress.yaml --kubeconfig ./test/kubeconfig.yaml
	sleep 2
	kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=160s --kubeconfig ./test/kubeconfig.yaml
