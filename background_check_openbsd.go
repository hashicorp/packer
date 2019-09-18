package main

import (
	"fmt"
)

func openTTY() (packer.TTY, error) {
	return false, fmt.Errorf("cannot determine if process is backgrounded in " +
		"openbsd")
}
