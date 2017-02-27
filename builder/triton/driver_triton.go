package triton

import (
	"errors"
	"strings"
	"time"

	"github.com/joyent/gosdc/cloudapi"
	"github.com/mitchellh/packer/packer"
)

type driverTriton struct {
	client *cloudapi.Client
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
	opts := cloudapi.CreateImageFromMachineOpts{
		Machine:     machineId,
		Name:        config.ImageName,
		Version:     config.ImageVersion,
		Description: config.ImageDescription,
		Homepage:    config.ImageHomepage,
		EULA:        config.ImageEULA,
		ACL:         config.ImageACL,
		Tags:        config.ImageTags,
	}

	image, err := d.client.CreateImageFromMachine(opts)
	if err != nil {
		return "", err
	}

	return image.Id, err
}

func (d *driverTriton) CreateMachine(config Config) (string, error) {
	opts := cloudapi.CreateMachineOpts{
		Package:         config.MachinePackage,
		Image:           config.MachineImage,
		Networks:        config.MachineNetworks,
		Metadata:        config.MachineMetadata,
		Tags:            config.MachineTags,
		FirewallEnabled: config.MachineFirewallEnabled,
	}

	if config.MachineName == "" {
		// If not supplied generate a name for the source VM: "packer-builder-[image_name]".
		// The version is not used because it can contain characters invalid for a VM name.
		opts.Name = "packer-builder-" + config.ImageName
	} else {
		opts.Name = config.MachineName
	}

	machine, err := d.client.CreateMachine(opts)
	if err != nil {
		return "", err
	}

	return machine.Id, nil
}

func (d *driverTriton) DeleteImage(imageId string) error {
	return d.client.DeleteImage(imageId)
}

func (d *driverTriton) DeleteMachine(machineId string) error {
	return d.client.DeleteMachine(machineId)
}

func (d *driverTriton) GetMachine(machineId string) (string, error) {
	machine, err := d.client.GetMachine(machineId)
	if err != nil {
		return "", err
	}

	return machine.PrimaryIP, nil
}

func (d *driverTriton) StopMachine(machineId string) error {
	return d.client.StopMachine(machineId)
}

// waitForMachineState uses the supplied client to wait for the state of
// the machine with the given ID to reach the state described in state.
// If timeout is reached before the machine reaches the required state, an
// error is returned. If the machine reaches the target state within the
// timeout, nil is returned.
func (d *driverTriton) WaitForMachineState(machineId string, state string, timeout time.Duration) error {
	return waitFor(
		func() (bool, error) {
			machine, err := d.client.GetMachine(machineId)
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
			machine, err := d.client.GetMachine(machineId)
			if err != nil {
				//TODO(jen20): is there a better way here than searching strings?
				if strings.Contains(err.Error(), "410") || strings.Contains(err.Error(), "404") {
					return true, nil
				}
			}

			if machine != nil {
				return false, nil
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
			image, err := d.client.GetImage(imageId)
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
