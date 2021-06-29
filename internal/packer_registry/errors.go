package packer_registry

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	_ = iota
	NonRegistryEnabled
	InvalidHCPConfig
)

type ClientError struct {
	StatusCode uint
	Err        error
}

func (c *ClientError) Error() string {
	return fmt.Sprintf("status %d: err %v", c.StatusCode, c.Err)
}

func NewNonRegistryEnabledError() error {
	return &ClientError{
		StatusCode: NonRegistryEnabled,
		Err:        errors.New("no Packer registry configuration found"),
	}
}

func IsNonRegistryEnabledError(err error) bool {
	clientErr, ok := err.(*ClientError)
	if !ok {
		return false
	}
	return clientErr.StatusCode != NonRegistryEnabled
}

func checkErrorCode(err error, code codes.Code) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	return st.Code() == code

}
