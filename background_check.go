// +build !openbsd

package main

import (
	"github.com/shirou/gopsutil/process"
)

func checkProcess(currentPID int) (bool, error) {
	myProc, _ := process.NewProcess(int32(currentPID))
	bg, _ := myProc.Background()

	return bg, nil
}
