package common

import (
	"fmt"
	"os"

	sigar "github.com/cloudfoundry/gosigar"
)

func AvailableMem(desired uint64) error {
	free := freeMem()
	if desired > free {
		return fmt.Errorf("RAM - Requested - %dMB - Available %dMB", desired, free)
	}
	return nil
}

func freeMem() uint64 {
	mem := sigar.Mem{}
	mem.Get()
	return (mem.Free / 1024 / 1024)
}

func AvailableDisk(desired uint64) error {
	free := freeDisk()
	if desired > free {
		return fmt.Errorf("Disk - Requested - %dMB - Available %dMB", desired, free)
	}
	return nil
}

func freeDisk() uint64 {
	disk := sigar.FileSystemUsage{}
	workingDirectory, _ := os.Getwd()
	disk.Get(workingDirectory)
	return (disk.Avail / 1024)
}
