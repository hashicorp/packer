// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build darwin || freebsd || linux || netbsd || openbsd || solaris
// +build darwin freebsd linux netbsd openbsd solaris

package command

const (
	defaultConfigDir = ".packer.d"
)
