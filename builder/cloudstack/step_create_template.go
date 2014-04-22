package cloudstack

import (
	"errors"
	"fmt"
	"github.com/mindjiver/gopherstack"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepCreateTemplate struct{}

func (s *stepCreateTemplate) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*gopherstack.CloudstackClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	vmid := state.Get("virtual_machine_id").(string)

	ui.Say(fmt.Sprintf("Creating template: %v", c.TemplateName))

	// get the volume id for the system volume for Virtual Machine 'id'
	response, err := client.ListVolumes(vmid)
	if err != nil {
		err := fmt.Errorf("Error creating template: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// always use the first volume when creating a template
	volumeId := response.Listvolumesresponse.Volume[0].ID
	response2, err := client.CreateTemplate(c.TemplateDisplayText, c.TemplateName,
		volumeId, c.TemplateOSId)
	if err != nil {
		err := fmt.Errorf("Error creating template: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting for template to be saved...")
	jobid := response2.Createtemplateresponse.Jobid
	err = client.WaitForAsyncJob(jobid, c.stateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for template to complete: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Looking up template ID for template: %s", c.TemplateName)
	response3, err := client.ListTemplates(c.TemplateName, "self")
	if err != nil {
		err := fmt.Errorf("Error looking up template ID: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Since if we create a template we should only have one with
	// that name, so we use the first response.
	template := response3.Listtemplatesresponse.Template[0].Name
	templateId := response3.Listtemplatesresponse.Template[0].ID

	if template != c.TemplateName {
		err := errors.New("Couldn't find template created. Bug?")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("template_name", template)
	state.Put("template_id", templateId)

	return multistep.ActionContinue
}

func (s *stepCreateTemplate) Cleanup(state multistep.StateBag) {
	// no cleanup
}
