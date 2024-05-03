// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP

// Package user contains ...
package main

import (
	"context"
	"github.com/naga2HPE/qt-test-application/internal/pkg/config"
	"github.com/naga2HPE/qt-test-application/internal/pkg/opentracing"
	"github.com/naga2HPE/qt-test-application/internal/pkg/user"
	logger "github.com/sirupsen/logrus"
	"log"
	"os"
	"github.com/grafana/pyroscope-go"
)

/*
package name    : user
project         : qt-test-application
*/

const serviceName = "user-service"

func main() {
	// read the config from .env file
	logger.SetFormatter(&logger.JSONFormatter{})
	logger.SetReportCaller(true)
	logger.SetLevel(logger.DebugLevel)
	logger.SetOutput(os.Stdout)
	logger.Infof("Server Starting...")

	// Get ApplicationName and ServerAddress from environmental variables
	applicationName := os.Getenv("APPLICATION_NAME")
	serverAddress := os.Getenv("PYROSCOPE_SERVER_ADDRESS")

	pyroscope.Start(pyroscope.Config{
		ApplicationName: applicationName,
		ServerAddress:   serverAddress,
		Logger:          pyroscope.StandardLogger,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		  },	  
	})

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

	// user.InitDB(config)
	user.SetupServer(config)

}
