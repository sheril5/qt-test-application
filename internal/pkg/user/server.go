// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP

// Package user contains ...
package user

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/naga2HPE/qt-test-application/internal/pkg/config"
	"github.com/naga2HPE/qt-test-application/internal/pkg/datastore"
	"github.com/naga2HPE/qt-test-application/internal/pkg/utils"
	"github.com/rs/cors"
	logger "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

/*
package name    : user
project         : qt-test-application
*/

const serviceName = "user-service"

var (
	db     datastore.DB
	srv    *http.Server
	tracer trace.Tracer
)

func SetupServer(configurations *config.ServiceConfigurations) {
	tracer = otel.Tracer(serviceName)
	router := mux.NewRouter()
	router.HandleFunc("/users", createUser).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/users/{userID}", getUser).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/users/{userID}", updateUser).Methods(http.MethodPut, http.MethodOptions)
	router.Use(utils.LoggingMW)
	router.Use(otelmux.Middleware(serviceName))
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost},
	})

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	srv = &http.Server{
		Addr:    configurations.UserURL,
		Handler: c.Handler(router),
	}

	log.Printf("User service running at: %s", configurations.UserURL)
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to setup http user: %v", err)
		}
	}()

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

type user struct {
	ID       int64  `json:"id" validate:"-"`
	UserName string `json:"user_name" validate:"required"`
	Account  string `json:"account" validate:"required"`
	Amount   int
}

type paymentData struct {
	Amount int `json:"amount" validate:"required"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u user
	if err := utils.ReadBody(w, r, &u); err != nil {
		return
	}

	_, span := tracer.Start(r.Context(), "create user")
	defer span.End()

	id := time.Now().UnixMilli()
	// id, err := db.InsertOne(ctx, datastore.InsertParams{
	// 	Query: `INSERT INTO USERS(USER_NAME, ACCOUNT) VALUES (?, ?)`,
	// 	Vars:  []interface{}{u.UserName, u.Account},
	// })
	// if err != nil {
	// 	utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("create user error: %w", err))
	// 	return
	// }
	logger.Infof("user ID :%s", id)
	u.ID = id
	utils.WriteResponse(w, http.StatusCreated, u)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	// var u user

	_, span := tracer.Start(r.Context(), "get user")
	defer span.End()
	span.SetAttributes(attribute.String("userID", userID))

	// if err := db.SelectOne(ctx, datastore.SelectParams{
	// 	Query:   `select ID, USER_NAME, ACCOUNT, AMOUNT from USERS where ID = ?`,
	// 	Filters: []interface{}{userID},
	// 	Result:  []interface{}{&u.ID, &u.UserName, &u.Account, &u.Amount},
	// }); err != nil {
	// 	utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("get user error: %w", err))
	// 	return
	// }

	utils.WriteResponse(w, http.StatusOK, user{ID: 1234, UserName: "JAD", Account: "jad", Amount: 1000})
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	var data paymentData
	if err := utils.ReadBody(w, r, &data); err != nil {
		return
	}

	_, span := tracer.Start(r.Context(), "update user amount")
	defer span.End()
	span.SetAttributes(attribute.String("userID", userID))
	// if err := db.UpdateOne(ctx, datastore.UpdateParams{
	// 	Query: `update USERS set AMOUNT = AMOUNT + ? where ID = ?`,
	// 	Vars:  []interface{}{data.Amount, userID},
	// }); err != nil {
	// 	utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("get user error: %w", err))
	// 	return
	// }

	w.WriteHeader(http.StatusOK)
}
