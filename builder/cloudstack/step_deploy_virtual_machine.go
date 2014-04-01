package cloudstack

import (
	"fmt"
	"github.com/mindjiver/gopherstack"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
)

type stepDeployVirtualMachine struct {
	id string
}

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort string
}

func (s *stepDeployVirtualMachine) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*gopherstack.CloudStackClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	sshKeyName := state.Get("ssh_key_name").(string)

	ui.Say("Creating virtual machine...")

	// Some random virtual machine name as it's temporary
	displayName := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Massage any userData that we wish to send to the virtual
	// machine to help it boot properly.
	processTemplatedUserdata(state)
	userData := state.Get("user_data").(string)

	// Create the virtual machine based on configuration
	response, err := client.DeployVirtualMachine(c.ServiceOfferingId,
		c.TemplateId, c.ZoneId, "", c.DiskOfferingId, displayName,
		c.NetworkIds, sshKeyName, "", userData, c.Hypervisor)

	if err != nil {
		err := fmt.Errorf("Error deploying virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Unpack the async jobid and wait for it
	vmid := response.Deployvirtualmachineresponse.ID
	jobid := response.Deployvirtualmachineresponse.Jobid
	client.WaitForAsyncJob(jobid, c.stateTimeout)

	// We use this in cleanup
	s.id = vmid

	// Store the virtual machine id for later use
	state.Put("virtual_machine_id", vmid)

	return multistep.ActionContinue
}

func (s *stepDeployVirtualMachine) Cleanup(state multistep.StateBag) {
	// If the virtual machine id isn't there, we probably never created it
	if s.id == "" {
		return
	}

	client := state.Get("client").(*gopherstack.CloudStackClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)

	// Destroy the virtual machine we just created
	ui.Say("Destroying virtual machine...")

	response, err := client.DestroyVirtualMachine(s.id)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying virtual machine. Please destroy it manually."))
	}
	jobid := response.Destroyvirtualmachineresponse.Jobid
	client.WaitForAsyncJob(jobid, c.stateTimeout)
}

func processTemplatedUserdata(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)

	// If there is no userdata to process we just save back an
	// empty string.
	if c.UserData == "" {
		state.Put("user_data", "")
		return multistep.ActionContinue
	}

	httpIP := state.Get("http_ip").(string)
	httpPort := state.Get("http_port").(string)

	tplData := &bootCommandTemplateData{
		httpIP,
		httpPort,
	}

	userData, err := c.tpl.Process(c.UserData, tplData)
	if err != nil {
		err := fmt.Errorf("Error preparing boot command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("user_data", userData)
	return multistep.ActionContinue
}
