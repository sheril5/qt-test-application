// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP

// Package order contains ...
package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/naga2HPE/qt-test-application/internal/pkg/config"
	"github.com/naga2HPE/qt-test-application/internal/pkg/datastore"
	"github.com/naga2HPE/qt-test-application/internal/pkg/utils"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
)

/*
package name    : order
project         : qt-test-application
*/

const serviceName = "order-service"

var (
	db      datastore.DB
	srv     *http.Server
	userUrl string
	tracer  trace.Tracer
)

func SetupServer(cnf *config.ServiceConfigurations) {
	tracer = otel.Tracer(serviceName)

	router := mux.NewRouter()
	router.HandleFunc("/orders", createOrder).Methods(http.MethodPost, http.MethodOptions)
	router.Use(utils.LoggingMW)
	router.Use(otelmux.Middleware(serviceName))
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost},
	})

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	userUrl = cnf.UserURL

	srv = &http.Server{
		Addr:    cnf.OrderURL,
		Handler: c.Handler(router),
	}

	log.Printf("Order service running at: %s", cnf.OrderURL)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to setup http server: %v", err)
	}

	<-sigint
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP user shutdown failed")
	}
}

func InitDB(cnf *config.ServiceConfigurations) {
	var err error
	if db, err = datastore.New(cnf); err != nil {
		log.Fatalf("failed to initialize db: %v", err)
	}
}

type orderData struct {
	ID          int64  `json:"id"`
	UserID      int    `json:"user_id" validate:"required"`
	ProductName string `json:"product_name" validate:"required"`
	Price       int    `json:"price" validate:"required"`
}

type user struct {
	ID       int64  `json:"id"`
	UserName string `json:"user_name"`
	Account  string `json:"account"`
	Amount   int
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var request orderData
	if err := utils.ReadBody(w, r, &request); err != nil {
		return
	}

	// get user details from user service
	url := fmt.Sprintf("http://%s/users/%d", userUrl, request.UserID)
	userResponse, err := utils.SendRequest(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		log.Printf("%v", err)
		utils.WriteResponse(w, http.StatusInternalServerError, err)
		return
	}

	b, err := ioutil.ReadAll(userResponse.Body)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer userResponse.Body.Close()

	if userResponse.StatusCode != http.StatusOK {
		utils.WriteErrorResponse(w, userResponse.StatusCode, fmt.Errorf("payment failed. got response: %s", b))
		return
	}

	var user user
	if err := json.Unmarshal(b, &user); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// basic check for the user balance
	if user.Amount < request.Price {
		utils.WriteErrorResponse(w, http.StatusUnprocessableEntity, fmt.Errorf("insufficient balance. add %d more amount to account", request.Price-user.Amount))
		return
	}

	// insert the order into order table
	ctx, insertSpan := tracer.Start(r.Context(), "insert order")
	id, err := db.InsertOne(ctx, datastore.InsertParams{
		Query: `insert into ORDERS(ACCOUNT, PRODUCT_NAME, PRICE, ORDER_STATUS) VALUES (?,?,?, ?)`,
		Vars:  []interface{}{user.Account, request.ProductName, request.Price, "SUCCESS"},
	})
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err)
		insertSpan.End()
		return
	}
	insertSpan.End()

	// update the pending amount in user table
	ctx, updateSpan := tracer.Start(r.Context(), "update user amount")
	if err := db.UpdateOne(ctx, datastore.UpdateParams{
		Query: `update USERS set AMOUNT = AMOUNT - ? where ID = ?`,
		Vars:  []interface{}{request.Price, user.ID},
	}); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err)
		updateSpan.End()
		return
	}
	updateSpan.End()

	// send response
	response := request
	response.ID = id
	utils.WriteResponse(w, http.StatusCreated, response)
}
