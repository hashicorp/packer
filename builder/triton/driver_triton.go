package triton

import (
	"errors"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/joyent/triton-go"
)

type driverTriton struct {
	client *triton.Client
	ui     packer.Ui
}

func NewDriverTriton(ui packer.Ui, config Config) (Driver, error) {
	client, err := config.AccessConfig.CreateTritonClient()
	if err != nil {
		return nil, err
	}

	return &driverTriton{
		client: client,
		ui:     ui,
	}, nil
}

func (d *driverTriton) CreateImageFromMachine(machineId string, config Config) (string, error) {
	image, err := d.client.Images().CreateImageFromMachine(&triton.CreateImageFromMachineInput{
		MachineID:   machineId,
		Name:        config.ImageName,
		Version:     config.ImageVersion,
		Description: config.ImageDescription,
		HomePage:    config.ImageHomepage,
		EULA:        config.ImageEULA,
		ACL:         config.ImageACL,
		Tags:        config.ImageTags,
	})
	if err != nil {
		return "", err
	}

	return image.ID, err
}

func (d *driverTriton) CreateMachine(config Config) (string, error) {
	input := &triton.CreateMachineInput{
		Package:         config.MachinePackage,
		Image:           config.MachineImage,
		Metadata:        config.MachineMetadata,
		Tags:            config.MachineTags,
		FirewallEnabled: config.MachineFirewallEnabled,
	}

	if config.MachineName == "" {
		// If not supplied generate a name for the source VM: "packer-builder-[image_name]".
		// The version is not used because it can contain characters invalid for a VM name.
		input.Name = "packer-builder-" + config.ImageName
	} else {
		input.Name = config.MachineName
	}

	if len(config.MachineNetworks) > 0 {
		input.Networks = config.MachineNetworks
	}

	machine, err := d.client.Machines().CreateMachine(input)
	if err != nil {
		return "", err
	}

	return machine.ID, nil
}

func (d *driverTriton) DeleteImage(imageId string) error {
	return d.client.Images().DeleteImage(&triton.DeleteImageInput{
		ImageID: imageId,
	})
}

func (d *driverTriton) DeleteMachine(machineId string) error {
	return d.client.Machines().DeleteMachine(&triton.DeleteMachineInput{
		ID: machineId,
	})
}

func (d *driverTriton) GetMachineIP(machineId string) (string, error) {
	machine, err := d.client.Machines().GetMachine(&triton.GetMachineInput{
		ID: machineId,
	})
	if err != nil {
		return "", err
	}

	return machine.PrimaryIP, nil
}

func (d *driverTriton) StopMachine(machineId string) error {
	return d.client.Machines().StopMachine(&triton.StopMachineInput{
		MachineID: machineId,
	})
}

// waitForMachineState uses the supplied client to wait for the state of
// the machine with the given ID to reach the state described in state.
// If timeout is reached before the machine reaches the required state, an
// error is returned. If the machine reaches the target state within the
// timeout, nil is returned.
func (d *driverTriton) WaitForMachineState(machineId string, state string, timeout time.Duration) error {
	return waitFor(
		func() (bool, error) {
			machine, err := d.client.Machines().GetMachine(&triton.GetMachineInput{
				ID: machineId,
			})
			if machine == nil {
				return false, err
			}
			return machine.State == state, err
		},
		3*time.Second,
		timeout,
	)
}

// waitForMachineDeletion uses the supplied client to wait for the machine
// with the given ID to be deleted. It is expected that the API call to delete
// the machine has already been issued at this point.
func (d *driverTriton) WaitForMachineDeletion(machineId string, timeout time.Duration) error {
	return waitFor(
		func() (bool, error) {
			machine, err := d.client.Machines().GetMachine(&triton.GetMachineInput{
				ID: machineId,
			})
			if err != nil && triton.IsResourceNotFound(err) {
				return true, nil
			}

			if machine != nil {
				return machine.State == "deleted", nil
			}

			return false, err
		},
		3*time.Second,
		timeout,
	)
}

func (d *driverTriton) WaitForImageCreation(imageId string, timeout time.Duration) error {
	return waitFor(
		func() (bool, error) {
			image, err := d.client.Images().GetImage(&triton.GetImageInput{
				ImageID: imageId,
			})
			if image == nil {
				return false, err
			}
			return image.OS != "", err
		},
		3*time.Second,
		timeout,
	)
}

func waitFor(f func() (bool, error), every, timeout time.Duration) error {
	start := time.Now()

	for time.Since(start) <= timeout {
		stop, err := f()
		if err != nil {
			return err
		}

		if stop {
			return nil
		}

		time.Sleep(every)
	}

	return errors.New("Timed out while waiting for resource change")
}
