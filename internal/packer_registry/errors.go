package packer_registry

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
)

const (
	_ = iota
	InvalidClientConfig
	NonRegistryEnabled
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
	var clientErr *ClientError
	if errors.As(err, &clientErr) {
		return clientErr.StatusCode == NonRegistryEnabled
	}

	return false
}

func checkErrorCode(err error, code codes.Code) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), fmt.Sprintf("Code:%d", code))

}
