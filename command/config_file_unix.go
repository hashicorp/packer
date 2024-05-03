// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build darwin || freebsd || linux || netbsd || openbsd || solaris
// +build darwin freebsd linux netbsd openbsd solaris

package command

const (
	defaultConfigDir = ".packer.d"
)
