// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build windows
// +build windows

package wrappedreadline

// getWidth impl for other
func getWidth() int {
	return 0
}
