package googlecompute

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCreateInstance represents a Packer build step that creates GCE instances.
type StepCreateInstance struct {
	Debug bool

	instanceName string
}

// Run executes the Packer build step that creates a GCE instance.
func (s *StepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	sshPublicKey := state.Get("ssh_public_key").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating instance...")
	name := config.InstanceName

	errCh, err := driver.RunInstance(&InstanceConfig{
		Description: "New instance created by Packer",
		Image:       config.SourceImage,
		MachineType: config.MachineType,
		Metadata: map[string]string{
			"sshKeys": fmt.Sprintf("%s:%s", config.SSHUsername, sshPublicKey),
		},
		Name:    name,
		Network: config.Network,
		Tags:    config.Tags,
		Zone:    config.Zone,
	})

	if err == nil {
		ui.Message("Waiting for creation operation to complete...")
		select {
		case err = <-errCh:
		case <-time.After(config.stateTimeout):
			err = errors.New("time out while waiting for instance to create")
		}
	}

	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message("Instance has been created!")

	if s.Debug {
		if name != "" {
			ui.Message(fmt.Sprintf("Instance: %s started in %s", name, config.Zone))
		}
	}

	// Things succeeded, store the name so we can remove it later
	state.Put("instance_name", name)
	s.instanceName = name

	return multistep.ActionContinue
}

// Cleanup destroys the GCE instance created during the image creation process.
func (s *StepCreateInstance) Cleanup(state multistep.StateBag) {
	if s.instanceName == "" {
		return
	}

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting instance...")
	errCh, err := driver.DeleteInstance(config.Zone, s.instanceName)
	if err == nil {
		select {
		case err = <-errCh:
		case <-time.After(config.stateTimeout):
			err = errors.New("time out while waiting for instance to delete")
		}
	}

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting instance. Please delete it manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", s.instanceName, err))
	}

	s.instanceName = ""
	return
}
