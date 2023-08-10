// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP

// Package payment contains ...
package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/naga2HPE/qt-test-application/internal/pkg/config"
	"github.com/naga2HPE/qt-test-application/internal/pkg/utils"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"log"
	"net/http"
)

/*
package name    : payment
project         : qt-test-application
*/
const serviceName = "payment-service"

var (
	srv     *http.Server
	userUrl string
	tracer  trace.Tracer
)

func SetupServer(configurations *config.ServiceConfigurations) {
	tracer = otel.Tracer(serviceName)

	router := mux.NewRouter()
	router.HandleFunc("/payments/transfer/id/{userID}", transferAmount).Methods(http.MethodPut, http.MethodOptions)
	router.Use(utils.LoggingMW)
	router.Use(otelmux.Middleware(serviceName))
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost},
	})

	userUrl = configurations.UserURL

	srv = &http.Server{
		Addr:    configurations.PaymentURL,
		Handler: c.Handler(router),
	}

	log.Printf("Payment service running at: %s", configurations.PaymentURL)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to setup http server: %v", err)
	}
}

type paymentData struct {
	Amount int `json:"amount" validate:"required"`
}

func transferAmount(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "transfer amount")
	defer span.End()
	userID := mux.Vars(r)["userID"]
	var data paymentData
	if err := utils.ReadBody(w, r, &data); err != nil {
		return
	}

	payload, err := json.Marshal(data)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// send the request to user service
	url := fmt.Sprintf("http://%s/users/%s", userUrl, userID)
	resp, err := utils.SendRequest(ctx, http.MethodPut, url, payload)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("payment failed. got response: %s", b))
		return
	}

	utils.WriteResponse(w, http.StatusOK, data)
}
