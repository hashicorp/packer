package triton

import (
	"time"
)

type DriverMock struct {
	CreateImageFromMachineId  string
	CreateImageFromMachineErr error

	CreateMachineId  string
	CreateMachineErr error

	DeleteImageId  string
	DeleteImageErr error

	DeleteMachineId  string
	DeleteMachineErr error

	GetMachineErr error

	StopMachineId  string
	StopMachineErr error

	WaitForImageCreationErr error

	WaitForMachineDeletionErr error

	WaitForMachineStateErr error
}

func (d *DriverMock) CreateImageFromMachine(machineId string, config Config) (string, error) {
	if d.CreateImageFromMachineErr != nil {
		return "", d.CreateImageFromMachineErr
	}

	d.CreateImageFromMachineId = config.ImageName

	return d.CreateImageFromMachineId, nil
}

func (d *DriverMock) CreateMachine(config Config) (string, error) {
	if d.CreateMachineErr != nil {
		return "", d.CreateMachineErr
	}

	d.CreateMachineId = config.MachineName

	return d.CreateMachineId, nil
}

func (d *DriverMock) DeleteImage(imageId string) error {
	if d.DeleteImageErr != nil {
		return d.DeleteImageErr
	}

	d.DeleteImageId = imageId

	return nil
}

func (d *DriverMock) DeleteMachine(machineId string) error {
	if d.DeleteMachineErr != nil {
		return d.DeleteMachineErr
	}

	d.DeleteMachineId = machineId

	return nil
}

func (d *DriverMock) GetMachine(machineId string) (string, error) {
	if d.GetMachineErr != nil {
		return "", d.GetMachineErr
	}

	return "ip", nil
}

func (d *DriverMock) StopMachine(machineId string) error {
	d.StopMachineId = machineId

	return d.StopMachineErr
}

func (d *DriverMock) WaitForImageCreation(machineId string, timeout time.Duration) error {
	return d.WaitForImageCreationErr
}

func (d *DriverMock) WaitForMachineDeletion(machineId string, timeout time.Duration) error {
	return d.WaitForMachineDeletionErr
}

func (d *DriverMock) WaitForMachineState(machineId string, state string, timeout time.Duration) error {
	return d.WaitForMachineStateErr
}
