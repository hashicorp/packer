package cloudstack

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/multistep"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

// userDataTemplateData represents variables for user_data interpolation
type userDataTemplateData struct {
	HTTPIP   string
	HTTPPort uint
}

// stepCreateInstance represents a Packer build step that creates CloudStack instances.
type stepCreateInstance struct {
	Ctx interpolate.Context
}

// Run executes the Packer build step that creates a CloudStack instance.
func (s *stepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating instance...")

	// Create a new parameter struct.
	p := client.VirtualMachine.NewDeployVirtualMachineParams(
		config.ServiceOffering,
		config.instanceSource,
		config.Zone,
	)

	// Configure the instance.
	p.SetName(config.InstanceName)
	p.SetDisplayname("Created by Packer")

	if config.Keypair != "" {
		p.SetKeypair(config.Keypair)
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
		httpPort := state.Get("http_port").(uint)
		hostIp, err := hostIP()
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.Ctx.Data = &userDataTemplateData{
			hostIp,
			httpPort,
		}

		renderedUserData, err := interpolate.Render(config.UserData, &s.Ctx)
		if err != nil {
			err := fmt.Errorf("Error rendering user_data: %s", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ud, err := getUserData(renderedUserData, config.HTTPGetOnly)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		p.SetUserdata(ud)
	}

	// Create the new instance.
	instance, err := client.VirtualMachine.DeployVirtualMachine(p)
	if err != nil {
		ui.Error(fmt.Sprintf("Error creating new instance %s: %s", config.InstanceName, err))
		return multistep.ActionHalt
	}

	ui.Message("Instance has been created!")

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
		config.hostAddress = instance.Nic[0].Ipaddress
	}

	// Store the instance ID so we can remove it later.
	state.Put("instance_id", instance.Id)

	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	ui := state.Get("ui").(packer.Ui)

	instanceID, ok := state.Get("instance_id").(string)
	if !ok || instanceID == "" {
		return
	}

	// Create a new parameter struct.
	p := client.VirtualMachine.NewDestroyVirtualMachineParams(instanceID)

	// Set expunge so the instance is completely removed
	p.SetExpunge(true)

	ui.Say("Deleting instance...")
	if _, err := client.VirtualMachine.DestroyVirtualMachine(p); err != nil {
		// This is a very poor way to be told the ID does no longer exist :(
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", instanceID)) {
			return
		}

		ui.Error(fmt.Sprintf("Error destroying instance: %s", err))
	}

	ui.Message("Instance has been deleted!")

	return
}

// getUserData returns the user data as a base64 encoded string.
func getUserData(userData string, httpGETOnly bool) (string, error) {
	ud := base64.StdEncoding.EncodeToString([]byte(userData))

	// deployVirtualMachine uses POST by default, so max userdata is 32K
	maxUD := 32768

	if httpGETOnly {
		// deployVirtualMachine using GET instead, so max userdata is 2K
		maxUD = 2048
	}

	if len(ud) > maxUD {
		return "", fmt.Errorf(
			"The supplied user_data contains %d bytes after encoding, "+
				"this exeeds the limit of %d bytes", len(ud), maxUD)
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
