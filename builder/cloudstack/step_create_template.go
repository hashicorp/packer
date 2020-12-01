package cloudstack

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepCreateTemplate struct{}

func (s *stepCreateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("Creating template: %s", config.TemplateName))

	// Retrieve the instance ID from the previously saved state.
	instanceID, ok := state.Get("instance_id").(string)
	if !ok || instanceID == "" {
		err := fmt.Errorf("Could not retrieve instance_id from state!")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Create a new parameter struct.
	p := client.Template.NewCreateTemplateParams(
		config.TemplateDisplayText,
		config.TemplateName,
		config.TemplateOS,
	)

	// Configure the template according to the supplied config.
	p.SetIsfeatured(config.TemplateFeatured)
	p.SetIspublic(config.TemplatePublic)
	p.SetIsdynamicallyscalable(config.TemplateScalable)
	p.SetPasswordenabled(config.TemplatePasswordEnabled)
	p.SetRequireshvm(config.TemplateRequiresHVM)

	if config.Project != "" {
		p.SetProjectid(config.Project)
	}

	if config.TemplateTag != "" {
		p.SetTemplatetag(config.TemplateTag)
	}

	ui.Message("Retrieving the ROOT volume ID...")
	volumeID, err := getRootVolumeID(client, instanceID, config.Project)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the volume ID from which to create the template.
	p.SetVolumeid(volumeID)

	ui.Message("Creating the new template...")
	template, err := client.Template.CreateTemplate(p)
	if err != nil {
		err := fmt.Errorf("Error creating the new template %s: %s", config.TemplateName, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// This is kind of nasty, but it appears to be needed to prevent corrupt templates.
	// When CloudStack says the template creation is done and you then delete the source
	// volume shortly after, it seems to corrupt the newly created template. Giving it an
	// additional 30 seconds to really finish up, seem to prevent that from happening.
	time.Sleep(30 * time.Second)

	ui.Message("Template has been created!")

	// Store the template.
	state.Put("template", template)

	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepCreateTemplate) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}

func getRootVolumeID(client *cloudstack.CloudStackClient, instanceID, projectID string) (string, error) {
	// Retrieve the virtual machine object.
	p := client.Volume.NewListVolumesParams()

	// Set the type and virtual machine ID
	p.SetType("ROOT")
	p.SetVirtualmachineid(instanceID)
	if projectID != "" {
		p.SetProjectid(projectID)
	}

	volumes, err := client.Volume.ListVolumes(p)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve ROOT volume: %s", err)
	}
	if volumes.Count != 1 {
		return "", fmt.Errorf("Could not find ROOT disk of instance %s", instanceID)
	}

	return volumes.Volumes[0].Id, nil
}
