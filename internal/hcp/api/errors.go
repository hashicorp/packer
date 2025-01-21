// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"fmt"
	"regexp"
	"strconv"

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

var errCodeRegex = regexp.MustCompilePOSIX(`[Cc]ode"?:([0-9]+)`)

// CheckErrorCode checks the error string for err for some code and returns true
// if the code is found. Ideally this function should use status.FromError
// https://pkg.go.dev/google.golang.org/grpc/status#pkg-functions but that
// doesn't appear to work for all of the Cloud Packer Service response errors.
func CheckErrorCode(err error, code codes.Code) bool {
	if err == nil {
		return false
	}

	// If the error string doesn't match the code we're looking for, we
	// can ignore it and return false immediately.
	matches := errCodeRegex.FindStringSubmatch(err.Error())
	if len(matches) == 0 {
		return false
	}

	// Safe to ignore the error here since the regex's submatch is always a
	// valid integer given the format ([0-9]+)
	errCode, _ := strconv.Atoi(matches[1])
	return errCode == int(code)
}
