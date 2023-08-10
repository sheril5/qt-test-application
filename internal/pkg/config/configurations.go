// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP

// Package config contains ...
package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/naga2HPE/qt-test-application/internal/pkg/gerrors"
)

/*
package name    : config
project         : qt-test-application
*/

type ServiceConfigurations struct {
	LogLevel     string `envconfig:"LOG_LEVEL" default:"info"`
	UserURL      string `envconfig:"USER_URL" default:"localhost:8081"`
	PaymentURL   string `envconfig:"PAYMENT_URL" default:"localhost:8082"`
	OrderURL     string `envconfig:"ORDER_URL" default:"localhost:8083"`
	SqlUser      string `envconfig:"SQL_USER" default:"root"`
	SqlPassword  string `envconfig:"SQL_PASSWORD" default:"password"`
	SqlHost      string `envconfig:"SQL_HOST" default:"localhost:3306"`
	SqlDB        string `envconfig:"SQL_DB" default:"signoz"`
	Collector    string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4317"`
	InsecureMode string `envconfig:"INSECURE_MODE" default:"true"`

	HeaderReadTimeout int
}

func GetServiceConfigurations() (serviceConf *ServiceConfigurations, err error) {
	serviceConf = &ServiceConfigurations{}
	if err = envconfig.Process("", serviceConf); err != nil {
		return nil, gerrors.NewFromError(gerrors.ServiceSetup, err)
	}
	return
}
