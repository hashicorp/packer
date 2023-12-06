// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !solaris
// +build !solaris

package main

import (
	"github.com/mattn/go-tty"
)

var openTTY = tty.Open
