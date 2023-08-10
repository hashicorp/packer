// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"fmt"
)

func checkProcess(currentPID int) (bool, error) {
	return false, fmt.Errorf("cannot determine if process is backgrounded in " +
		"openbsd")
}
