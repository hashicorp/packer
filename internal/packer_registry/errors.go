package packer_registry

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	Unknown = iota
	UnsetClient
)

type HCPClientError struct {
	StatusCode uint
	Err        error
}

func (c *HCPClientError) Error() string {
	return fmt.Sprintf("status %d: err %v", c.StatusCode, c.Err)
}

func (c *HCPClientError) Fatal() bool {
	return c.StatusCode != UnsetClient
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
