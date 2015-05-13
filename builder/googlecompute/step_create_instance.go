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
}

func (config *Config) getImage() Image {
	project := config.ProjectId
	if config.SourceImageProjectId != "" {
		project = config.SourceImageProjectId
	}
	return Image{Name: config.SourceImage, ProjectId: project}
}

func (config *Config) getInstanceMetadata(sshPublicKey string) map[string]string {
	instanceMetadata := make(map[string]string)

	// Copy metadata from config
	for k, v := range config.Metadata {
		instanceMetadata[k] = v
	}

	// Merge any existing ssh keys with our public key
	sshMetaKey := "sshKeys"
	sshKeys := fmt.Sprintf("%s:%s", config.SSHUsername, sshPublicKey)
	if confSshKeys, exists := instanceMetadata[sshMetaKey]; exists {
		sshKeys = fmt.Sprintf("%s\n%s", sshKeys, confSshKeys)
	}
	instanceMetadata[sshMetaKey] = sshKeys

	return instanceMetadata
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
		DiskSizeGb:  config.DiskSizeGb,
		Image:       config.getImage(),
		MachineType: config.MachineType,
		Metadata:    config.getInstanceMetadata(sshPublicKey),
		Name:        name,
		Network:     config.Network,
		Tags:        config.Tags,
		Zone:        config.Zone,
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

	return multistep.ActionContinue
}

// Cleanup destroys the GCE instance created during the image creation process.
func (s *StepCreateInstance) Cleanup(state multistep.StateBag) {
	nameRaw, ok := state.GetOk("instance_name")
	if !ok {
		return
	}
	name := nameRaw.(string)
	if name == "" {
		return
	}

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting instance...")
	errCh, err := driver.DeleteInstance(config.Zone, name)
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
				"Error: %s", name, err))
	}

	ui.Message("Instance has been deleted!")
	state.Put("instance_name", "")

	// Deleting the instance does not remove the boot disk. This cleanup removes
	// the disk.
	ui.Say("Deleting disk...")
	errCh, err = driver.DeleteDisk(config.Zone, config.DiskName)
	if err == nil {
		select {
		case err = <-errCh:
		case <-time.After(config.stateTimeout):
			err = errors.New("time out while waiting for disk to delete")
		}
	}

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting disk. Please delete it manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", config.InstanceName, err))
	}

	ui.Message("Disk has been deleted!")

	return
}
