// +build !openbsd

package main

import (
	"fmt"

	"github.com/shirou/gopsutil/process"
)

func checkProcess(currentPID int) (bool, error) {
	myProc, err := process.NewProcess(int32(currentPID))
	if err != nil {
		return false, fmt.Errorf("Error figuring out Packer process info")
	}
	bg, _ := myProc.Background()

	return bg, nil
}
