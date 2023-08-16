// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !solaris
// +build !solaris

package main

import (
	"github.com/mattn/go-tty"
)

var openTTY = tty.Open
