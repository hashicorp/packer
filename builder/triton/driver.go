package triton

import (
	"time"
)

type Driver interface {
	CreateImageFromMachine(machineId string, config Config) (string, error)
	CreateMachine(config Config) (string, error)
	DeleteImage(imageId string) error
	DeleteMachine(machineId string) error
	GetMachine(machineId string) (string, error)
	StopMachine(machineId string) error
	WaitForImageCreation(imageId string, timeout time.Duration) error
	WaitForMachineDeletion(machineId string, timeout time.Duration) error
	WaitForMachineState(machineId string, state string, timeout time.Duration) error
}
