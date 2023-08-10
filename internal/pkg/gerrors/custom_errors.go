// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP

// Package gerrors contains ...
package gerrors

/*
package name    : gerrors
project         : sample-http-user
*/

type ErrorCode string

const (
	InternalError        ErrorCode = "Internal Error"
	ServiceSetup         ErrorCode = "ServiceSetup"
	ValidationFailed     ErrorCode = "Validations Failed"
	BadRequest           ErrorCode = "Bad Request"
	NotFound             ErrorCode = "Not Found"
	LDAPClient           ErrorCode = "LDAP Client"
	TokenNotFound        ErrorCode = "Token NotFound"
	AuthenticationFailed ErrorCode = "Authentication Failed"

	InvalidDBConfig ErrorCode = "Invalid DB Configurations"
	InvalidInput    ErrorCode = "Bad Request"
)
