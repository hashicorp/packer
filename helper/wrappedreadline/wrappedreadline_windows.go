// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build windows
// +build windows

package wrappedreadline

// getWidth impl for other
func getWidth() int {
	return 0
}
