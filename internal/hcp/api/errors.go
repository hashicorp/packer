// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
)

const (
	_ = iota
	InvalidClientConfig
)

// ClientError represents a generic error for the Cloud Packer Service client.
type ClientError struct {
	StatusCode uint
	Err        error
}

// Error returns the string message for some ClientError.
func (c *ClientError) Error() string {
	return fmt.Sprintf("status %d: err %v", c.StatusCode, c.Err)
}

// CheckErrorCode checks the error string for err for some code and returns true
// if the code is found. Ideally this function should use status.FromError
// https://pkg.go.dev/google.golang.org/grpc/status#pkg-functions but that
// doesn't appear to work for all of the Cloud Packer Service response errors.
func CheckErrorCode(err error, code codes.Code) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), fmt.Sprintf("Code:%d", code))
}
