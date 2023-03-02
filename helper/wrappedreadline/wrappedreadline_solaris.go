// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wrappedreadline

// getWidth impl for Solaris
func getWidth() int {
	return 80
}

// get width of the terminal
func getWidthFd(stdoutFd int) int {
	return getWidth()
}
