package opc

import "fmt"

// OracleError details the parameters of an error returned from Oracle's API
type OracleError struct {
	StatusCode int
	Message    string
}

func (e OracleError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}
