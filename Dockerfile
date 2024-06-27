# Build the manager binary
FROM FROM ghcr.io/kube-tarian/helmrepo-supporting-tools/golang:1.22-alpine as builder

# Update and upgrade packages
RUN apk update && apk upgrade && apk add make

WORKDIR /workspace
# Copy the Go Modules manifests
COPY ./ ./

# Build
#RUN make proto-gen-all
RUN  make vendor
RUN  make build_user

CMD "h b m n b b n"

# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o builds/user cmd/user/main.go

FROM scratch
COPY --from=builder /workspace/builds/user .

CMD [ "./user" ]
