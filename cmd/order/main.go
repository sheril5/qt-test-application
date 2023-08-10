// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP

// Package order contains ...
package main

import (
	"context"
	"github.com/naga2HPE/qt-test-application/internal/pkg/config"
	"github.com/naga2HPE/qt-test-application/internal/pkg/opentracing"
	"github.com/naga2HPE/qt-test-application/internal/pkg/order"
	logger "github.com/sirupsen/logrus"
	"log"
	"os"
)

/*
package name    : order
project         : qt-test-application
*/

const serviceName = "order-service"

func main() {
	// read the config from .env file
	logger.SetFormatter(&logger.JSONFormatter{})
	logger.SetReportCaller(true)
	logger.SetLevel(logger.DebugLevel)
	logger.SetOutput(os.Stdout)
	logger.Infof("Server Starting...")

	config, err := config.GetServiceConfigurations()
	if err != nil {
		os.Exit(1)
	}

	// setup tracer
	tp := opentracing.Init(config, serviceName)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	order.InitDB(config)
	order.SetupServer(config)

}
