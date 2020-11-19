package googlecompute

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepCreateInstance represents a Packer build step that creates GCE instances.
type StepCreateInstance struct {
	Debug bool
}

func (c *Config) createInstanceMetadata(sourceImage *Image, sshPublicKey string) (map[string]string, error) {
	instanceMetadata := make(map[string]string)
	var err error
	var errs *packer.MultiError

	// Copy metadata from config.
	for k, v := range c.Metadata {
		instanceMetadata[k] = v
	}

	// Merge any existing ssh keys with our public key, unless there is no
	// supplied public key. This is possible if a private_key_file was
	// specified.
	if sshPublicKey != "" {
		sshMetaKey := "ssh-keys"
		sshPublicKey = strings.TrimSuffix(sshPublicKey, "\n")
		sshKeys := fmt.Sprintf("%s:%s %s", c.Comm.SSHUsername, sshPublicKey, c.Comm.SSHUsername)
		if confSshKeys, exists := instanceMetadata[sshMetaKey]; exists {
			sshKeys = fmt.Sprintf("%s\n%s", sshKeys, confSshKeys)
		}
		instanceMetadata[sshMetaKey] = sshKeys
	}

	startupScript := instanceMetadata[StartupScriptKey]
	if c.StartupScriptFile != "" {
		var content []byte
		content, err = ioutil.ReadFile(c.StartupScriptFile)
		if err != nil {
			return nil, err
		}
		startupScript = string(content)
	}
	instanceMetadata[StartupScriptKey] = startupScript

	// Wrap any found startup script with our own startup script wrapper.
	if startupScript != "" && c.WrapStartupScriptFile.True() {
		instanceMetadata[StartupScriptKey] = StartupScriptLinux
		instanceMetadata[StartupWrappedScriptKey] = startupScript
		instanceMetadata[StartupScriptStatusKey] = StartupScriptStatusNotDone
	}

	if sourceImage.IsWindows() {
		// Windows startup script support is not yet implemented so clear any script data and set status to done
		instanceMetadata[StartupScriptKey] = StartupScriptWindows
		instanceMetadata[StartupScriptStatusKey] = StartupScriptStatusDone
	}

	// If UseOSLogin is true, force `enable-oslogin` in metadata
	// In the event that `enable-oslogin` is not enabled at project level
	if c.UseOSLogin {
		instanceMetadata[EnableOSLoginKey] = "TRUE"
	}

	for key, value := range c.MetadataFiles {
		var content []byte
		content, err = ioutil.ReadFile(value)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
		instanceMetadata[key] = string(content)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return instanceMetadata, errs
	}
	return instanceMetadata, nil
}

func getImage(c *Config, d Driver) (*Image, error) {
	name := c.SourceImageFamily
	fromFamily := true
	if c.SourceImage != "" {
		name = c.SourceImage
		fromFamily = false
	}
	if len(c.SourceImageProjectId) == 0 {
		return d.GetImage(name, fromFamily)
	} else {
		return d.GetImageFromProjects(c.SourceImageProjectId, name, fromFamily)
	}
}

// Run executes the Packer build step that creates a GCE instance.
func (s *StepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	d := state.Get("driver").(Driver)

	ui := state.Get("ui").(packersdk.Ui)

	sourceImage, err := getImage(c, d)
	if err != nil {
		err := fmt.Errorf("Error getting source image for instance creation: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if c.EnableSecureBoot && !sourceImage.IsSecureBootCompatible() {
		err := fmt.Errorf("Image: %s is not secure boot compatible. Please set 'enable_secure_boot' to false or choose another source image.", sourceImage.Name)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Using image: %s", sourceImage.Name))

	if sourceImage.IsWindows() && c.Comm.Type == "winrm" && c.Comm.WinRMPassword == "" {
		state.Put("create_windows_password", true)
	}

	ui.Say("Creating instance...")
	name := c.InstanceName

	var errCh <-chan error
	var metadata map[string]string
	metadata, errs := c.createInstanceMetadata(sourceImage, string(c.Comm.SSHPublicKey))
	if errs != nil {
		state.Put("error", errs.Error())
		ui.Error(errs.Error())
		return multistep.ActionHalt
	}

	errCh, err = d.RunInstance(&InstanceConfig{
		AcceleratorType:              c.AcceleratorType,
		AcceleratorCount:             c.AcceleratorCount,
		Address:                      c.Address,
		Description:                  "New instance created by Packer",
		DisableDefaultServiceAccount: c.DisableDefaultServiceAccount,
		DiskSizeGb:                   c.DiskSizeGb,
		DiskType:                     c.DiskType,
		EnableSecureBoot:             c.EnableSecureBoot,
		EnableVtpm:                   c.EnableVtpm,
		EnableIntegrityMonitoring:    c.EnableIntegrityMonitoring,
		Image:                        sourceImage,
		Labels:                       c.Labels,
		MachineType:                  c.MachineType,
		Metadata:                     metadata,
		MinCpuPlatform:               c.MinCpuPlatform,
		Name:                         name,
		Network:                      c.Network,
		NetworkProjectId:             c.NetworkProjectId,
		OmitExternalIP:               c.OmitExternalIP,
		OnHostMaintenance:            c.OnHostMaintenance,
		Preemptible:                  c.Preemptible,
		Region:                       c.Region,
		ServiceAccountEmail:          c.ServiceAccountEmail,
		Scopes:                       c.Scopes,
		Subnetwork:                   c.Subnetwork,
		Tags:                         c.Tags,
		Zone:                         c.Zone,
	})

	if err == nil {
		ui.Message("Waiting for creation operation to complete...")
		select {
		case err = <-errCh:
		case <-time.After(c.StateTimeout):
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
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", name)

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
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting instance...")
	errCh, err := driver.DeleteInstance(config.Zone, name)
	if err == nil {
		select {
		case err = <-errCh:
		case <-time.After(config.StateTimeout):
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
		case <-time.After(config.StateTimeout):
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
