// +build !solaris

package main

import (
	"github.com/mattn/go-tty"
)

var openTTY = tty.Open
