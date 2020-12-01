package dtl

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/devtestlabs/mgmt/2018-09-15/dtl"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepCaptureImage struct {
	client              *AzureClient
	captureManagedImage func(ctx context.Context) error
	get                 func(client *AzureClient) *CaptureTemplate
	config              *Config
	say                 func(message string)
	error               func(e error)
}

func NewStepCaptureImage(client *AzureClient, ui packersdk.Ui, config *Config) *StepCaptureImage {
	var step = &StepCaptureImage{
		client: client,
		get: func(client *AzureClient) *CaptureTemplate {
			return client.Template
		},
		config: config,
		say: func(message string) {
			ui.Say(message)
		},
		error: func(e error) {
			ui.Error(e.Error())
		},
	}

	// step.captureVhd = step.captureImage
	step.captureManagedImage = step.captureImageFromVM

	return step
}

func (s *StepCaptureImage) captureImageFromVM(ctx context.Context) error {
	imageID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DevTestLab/labs/%s/virtualMachines/%s",
		s.config.ClientConfig.SubscriptionID,
		s.config.tmpResourceGroupName,
		s.config.LabName,
		s.config.tmpComputeName)

	customImageProperties := dtl.CustomImageProperties{}

	if s.config.OSType == constants.Target_Linux {
		customImageProperties = dtl.CustomImageProperties{
			VM: &dtl.CustomImagePropertiesFromVM{
				LinuxOsInfo: &dtl.LinuxOsInfo{
					LinuxOsState: dtl.DeprovisionRequested,
				},
				SourceVMID: &imageID,
			},
		}
	} else if s.config.OSType == constants.Target_Windows {
		customImageProperties = dtl.CustomImageProperties{
			VM: &dtl.CustomImagePropertiesFromVM{
				WindowsOsInfo: &dtl.WindowsOsInfo{
					WindowsOsState: dtl.SysprepRequested,
				},
				SourceVMID: &imageID,
			},
		}
	}

	customImage := &dtl.CustomImage{
		Name:                  &s.config.ManagedImageName,
		CustomImageProperties: &customImageProperties,
	}

	f, err := s.client.DtlCustomImageClient.CreateOrUpdate(ctx, s.config.LabResourceGroupName, s.config.LabName, s.config.ManagedImageName, *customImage)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, s.client.DtlCustomImageClient.Client)
	}
	if err != nil {
		s.say("Error from Capture Image")
		s.say(s.client.LastError.Error())
	}

	return err
}

func (s *StepCaptureImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Capturing image ...")

	var computeName = state.Get(constants.ArmComputeName).(string)
	var location = state.Get(constants.ArmLocation).(string)
	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)

	s.say(fmt.Sprintf(" -> Compute ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> Compute Name              : '%s'", computeName))
	s.say(fmt.Sprintf(" -> Compute Location          : '%s'", location))

	err := s.captureImageFromVM(ctx)

	if err != nil {
		s.error(err)
		state.Put(constants.Error, err)

		return multistep.ActionHalt
	}

	// HACK(chrboum): I do not like this.  The capture method should be returning this value
	// instead having to pass in another lambda.
	//
	// Having to resort to capturing the template via an inspector is hack, and once I can
	// resolve that I can cleanup this code too.  See the comments in azure_client.go for more
	// details.
	// [paulmey]: autorest.Future now has access to the last http.Response, but I'm not sure if
	// the body is still accessible.
	template := s.get(s.client)
	state.Put(constants.ArmCaptureTemplate, template)

	return multistep.ActionContinue
}

func (*StepCaptureImage) Cleanup(multistep.StateBag) {
}
