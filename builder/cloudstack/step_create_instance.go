package cloudstack

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

// userDataTemplateData represents variables for user_data interpolation
type userDataTemplateData struct {
	HTTPIP   string
	HTTPPort int
}

// stepCreateInstance represents a Packer build step that creates CloudStack instances.
type stepCreateInstance struct {
	Debug bool
	Ctx   interpolate.Context
}

// Run executes the Packer build step that creates a CloudStack instance.
func (s *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Creating instance...")

	// Create a new parameter struct.
	p := client.VirtualMachine.NewDeployVirtualMachineParams(
		config.ServiceOffering,
		state.Get("source").(string),
		config.Zone,
	)

	// Configure the instance.
	p.SetName(config.InstanceName)
	p.SetDisplayname(config.InstanceDisplayName)

	if len(config.Comm.SSHKeyPairName) != 0 {
		ui.Message(fmt.Sprintf("Using keypair: %s", config.Comm.SSHKeyPairName))
		p.SetKeypair(config.Comm.SSHKeyPairName)
	}

	if securitygroups, ok := state.GetOk("security_groups"); ok {
		p.SetSecuritygroupids(securitygroups.([]string))
	}

	// If we use an ISO, configure the disk offering.
	if config.SourceISO != "" {
		p.SetDiskofferingid(config.DiskOffering)
		p.SetHypervisor(config.Hypervisor)
	}

	// If we use a template, set the root disk size.
	if config.SourceTemplate != "" && config.DiskSize > 0 {
		p.SetRootdisksize(config.DiskSize)
	}

	// Retrieve the zone object.
	zone, _, err := client.Zone.GetZoneByID(config.Zone)
	if err != nil {
		err := fmt.Errorf("Failed to get zone %s by ID: %s", config.Zone, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if zone.Networktype == "Advanced" {
		// Set the network ID's.
		p.SetNetworkids([]string{config.Network})
	}

	// If there is a project supplied, set the project id.
	if config.Project != "" {
		p.SetProjectid(config.Project)
	}

	if config.UserData != "" {
		httpPort := state.Get("http_port").(int)
		httpIP, err := hostIP()
		if err != nil {
			err := fmt.Errorf("Failed to determine host IP: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		state.Put("http_ip", httpIP)

		s.Ctx.Data = &userDataTemplateData{
			httpIP,
			httpPort,
		}

		ud, err := s.generateUserData(config.UserData, config.HTTPGetOnly)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		p.SetUserdata(ud)
	}

	// Create the new instance.
	instance, err := client.VirtualMachine.DeployVirtualMachine(p)
	if err != nil {
		err := fmt.Errorf("Error creating new instance %s: %s", config.InstanceName, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message("Instance has been created!")
	ui.Message(fmt.Sprintf("Instance ID: %s", instance.Id))

	// In debug-mode, we output the password
	if s.Debug {
		ui.Message(fmt.Sprintf(
			"Password (since debug is enabled) \"%s\"", instance.Password))
	}

	// Set the auto generated password if a password was not explicitly configured.
	switch config.Comm.Type {
	case "ssh":
		if config.Comm.SSHPassword == "" {
			config.Comm.SSHPassword = instance.Password
		}
	case "winrm":
		if config.Comm.WinRMPassword == "" {
			config.Comm.WinRMPassword = instance.Password
		}
	}

	// Set the host address when using the local IP address to connect.
	if config.UseLocalIPAddress {
		state.Put("ipaddress", instance.Nic[0].Ipaddress)
	}

	// Store the instance ID so we can remove it later.
	state.Put("instance_id", instance.Id)

	// Set instance tags
	if len(config.Tags) > 0 {
		resourceID := []string{instance.Id}
		tp := client.Resourcetags.NewCreateTagsParams(resourceID, "UserVm", config.Tags)

		_, err = client.Resourcetags.CreateTags(tp)

		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	instanceID, ok := state.Get("instance_id").(string)
	if !ok || instanceID == "" {
		return
	}

	// Create a new parameter struct.
	p := client.VirtualMachine.NewDestroyVirtualMachineParams(instanceID)

	ui.Say("Deleting instance...")
	if _, err := client.VirtualMachine.DestroyVirtualMachine(p); err != nil {
		// This is a very poor way to be told the ID does no longer exist :(
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", instanceID)) {
			return
		}

		ui.Error(fmt.Sprintf("Error destroying instance. Please destroy it manually.\n\n"+
			"\tName: %s\n"+
			"\tError: %s", config.InstanceName, err))
		return
	}

	// We could expunge the VM while destroying it, but if the user doesn't have
	// rights that single call could error out leaving the VM running. So but
	// splitting these calls we make sure the VM is always deleted, even when the
	// expunge fails.
	if config.Expunge {
		// Create a new parameter struct.
		p := client.VirtualMachine.NewExpungeVirtualMachineParams(instanceID)

		ui.Say("Expunging instance...")
		if _, err := client.VirtualMachine.ExpungeVirtualMachine(p); err != nil {
			// This is a very poor way to be told the ID does no longer exist :(
			if strings.Contains(err.Error(), fmt.Sprintf(
				"Invalid parameter id value=%s due to incorrect long value format, "+
					"or entity does not exist", instanceID)) {
				return
			}

			ui.Error(fmt.Sprintf("Error expunging instance. Please expunge it manually.\n\n"+
				"\tName: %s\n"+
				"\tError: %s", config.InstanceName, err))
			return
		}
	}

	ui.Message("Instance has been deleted!")
	return
}

// generateUserData returns the user data as a base64 encoded string.
func (s *stepCreateInstance) generateUserData(userData string, httpGETOnly bool) (string, error) {
	renderedUserData, err := interpolate.Render(userData, &s.Ctx)
	if err != nil {
		return "", fmt.Errorf("Error rendering user_data: %s", err)
	}

	ud := base64.StdEncoding.EncodeToString([]byte(renderedUserData))

	// DeployVirtualMachine uses POST by default which allows 32K of
	// userdata. If using GET instead the userdata is limited to 2K.
	maxUD := 32768
	if httpGETOnly {
		maxUD = 2048
	}

	if len(ud) > maxUD {
		return "", fmt.Errorf(
			"The supplied user_data contains %d bytes after encoding, "+
				"this exceeds the limit of %d bytes", len(ud), maxUD)
	}

	return ud, nil
}

func hostIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("No host IP found")
}
