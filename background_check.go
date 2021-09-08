//go:build !openbsd
// +build !openbsd

package main

import (
	"fmt"

	"github.com/shirou/gopsutil/process"
)

func checkProcess(currentPID int) (bool, error) {
	myProc, err := process.NewProcess(int32(currentPID))
	if err != nil {
		return false, fmt.Errorf("Process check error: %s", err)
	}
	bg, err := myProc.Background()
	if err != nil {
		return bg, fmt.Errorf("Process background check error: %s", err)
	}

	return bg, nil
}
