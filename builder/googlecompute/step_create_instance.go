package googlecompute

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCreateInstance represents a Packer build step that creates GCE instances.
type StepCreateInstance struct {
	Debug bool
}

func (c *Config) createInstanceMetadata(sourceImage *Image, sshPublicKey string) (map[string]string, error) {
	instanceMetadata := make(map[string]string)
	var err error

	// Copy metadata from config.
	for k, v := range c.Metadata {
		instanceMetadata[k] = v
	}

	// Merge any existing ssh keys with our public key.
	sshMetaKey := "sshKeys"
	sshKeys := fmt.Sprintf("%s:%s", c.Comm.SSHUsername, sshPublicKey)
	if confSshKeys, exists := instanceMetadata[sshMetaKey]; exists {
		sshKeys = fmt.Sprintf("%s\n%s", sshKeys, confSshKeys)
	}
	instanceMetadata[sshMetaKey] = sshKeys

	// Wrap any startup script with our own startup script.
	if c.StartupScriptFile != "" {
		var content []byte
		content, err = ioutil.ReadFile(c.StartupScriptFile)
		instanceMetadata[StartupWrappedScriptKey] = string(content)
	} else if wrappedStartupScript, exists := instanceMetadata[StartupScriptKey]; exists {
		instanceMetadata[StartupWrappedScriptKey] = wrappedStartupScript
	}
	if sourceImage.IsWindows() {
		// Windows startup script support is not yet implemented.
		// Mark the startup script as done.
		instanceMetadata[StartupScriptKey] = StartupScriptWindows
		instanceMetadata[StartupScriptStatusKey] = StartupScriptStatusDone
	} else {
		instanceMetadata[StartupScriptKey] = StartupScriptLinux
		instanceMetadata[StartupScriptStatusKey] = StartupScriptStatusNotDone
	}

	// Adding the Sysprep script to enable Windows RM
	instanceMetadata[StartupWindowsSysprepKey] = StartupWinRMScript

	return instanceMetadata, err
}

func getImage(c *Config, d Driver) (*Image, error) {
	if c.SourceImageProjectId == "" {
		return d.GetImage(c.SourceImage)
	} else {
		return d.GetImageFromProject(c.SourceImageProjectId, c.SourceImage)
	}
}

// Run executes the Packer build step that creates a GCE instance.
func (s *StepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	d := state.Get("driver").(Driver)
	sshPublicKey := state.Get("ssh_public_key").(string)
	ui := state.Get("ui").(packer.Ui)

	sourceImage, err := getImage(c, d)
	if err != nil {
		err := fmt.Errorf("Error getting source image for instance creation: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Creating instance...")
	name := c.InstanceName

	var errCh <-chan error
	var metadata map[string]string
	metadata, err = c.createInstanceMetadata(sourceImage, sshPublicKey)
	errCh, err = d.RunInstance(&InstanceConfig{
		Address:             c.Address,
		Description:         "New instance created by Packer",
		DiskSizeGb:          c.DiskSizeGb,
		DiskType:            c.DiskType,
		Image:               sourceImage,
		MachineType:         c.MachineType,
		Metadata:            metadata,
		Name:                name,
		Network:             c.Network,
		OmitExternalIP:      c.OmitExternalIP,
		Preemptible:         c.Preemptible,
		Region:              c.Region,
		ServiceAccountEmail: c.Account.ClientEmail,
		Subnetwork:          c.Subnetwork,
		Tags:                c.Tags,
		Zone:                c.Zone,
	})

	if err == nil {
		ui.Message("Waiting for creation operation to complete...")
		select {
		case err = <-errCh:
		case <-time.After(c.stateTimeout):
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
			ui.Message(fmt.Sprintf("Instance: %s started in %s", name, c.Zone))
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
