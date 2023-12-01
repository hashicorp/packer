// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package registry

// ErrBuildAlreadyDone is the error returned by an HCP handler when a build cannot be started since it's already marked as DONE.
type ErrBuildAlreadyDone struct {
	Message string
}

// Error returns the message for the ErrBuildAlreadyDone type
func (b ErrBuildAlreadyDone) Error() string {
	return b.Message
}
