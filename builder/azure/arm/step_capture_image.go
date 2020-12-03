package arm

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepCaptureImage struct {
	client              *AzureClient
	generalizeVM        func(resourceGroupName, computeName string) error
	captureVhd          func(ctx context.Context, resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters) error
	captureManagedImage func(ctx context.Context, resourceGroupName string, computeName string, parameters *compute.Image) error
	get                 func(client *AzureClient) *CaptureTemplate
	say                 func(message string)
	error               func(e error)
}

func NewStepCaptureImage(client *AzureClient, ui packersdk.Ui) *StepCaptureImage {
	var step = &StepCaptureImage{
		client: client,
		get: func(client *AzureClient) *CaptureTemplate {
			return client.Template
		},
		say: func(message string) {
			ui.Say(message)
		},
		error: func(e error) {
			ui.Error(e.Error())
		},
	}

	step.generalizeVM = step.generalize
	step.captureVhd = step.captureImage
	step.captureManagedImage = step.captureImageFromVM

	return step
}

func (s *StepCaptureImage) generalize(resourceGroupName string, computeName string) error {
	_, err := s.client.Generalize(context.TODO(), resourceGroupName, computeName)
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepCaptureImage) captureImageFromVM(ctx context.Context, resourceGroupName string, imageName string, image *compute.Image) error {
	f, err := s.client.ImagesClient.CreateOrUpdate(ctx, resourceGroupName, imageName, *image)
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return f.WaitForCompletionRef(ctx, s.client.ImagesClient.Client)
}

func (s *StepCaptureImage) captureImage(ctx context.Context, resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters) error {
	f, err := s.client.VirtualMachinesClient.Capture(ctx, resourceGroupName, computeName, *parameters)
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return f.WaitForCompletionRef(ctx, s.client.VirtualMachinesClient.Client)
}

func (s *StepCaptureImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Capturing image ...")

	var computeName = state.Get(constants.ArmComputeName).(string)
	var location = state.Get(constants.ArmLocation).(string)
	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var vmCaptureParameters = state.Get(constants.ArmVirtualMachineCaptureParameters).(*compute.VirtualMachineCaptureParameters)
	var imageParameters = state.Get(constants.ArmImageParameters).(*compute.Image)

	var isManagedImage = state.Get(constants.ArmIsManagedImage).(bool)
	var targetManagedImageResourceGroupName = state.Get(constants.ArmManagedImageResourceGroupName).(string)
	var targetManagedImageName = state.Get(constants.ArmManagedImageName).(string)
	var targetManagedImageLocation = state.Get(constants.ArmLocation).(string)

	s.say(fmt.Sprintf(" -> Compute ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> Compute Name              : '%s'", computeName))
	s.say(fmt.Sprintf(" -> Compute Location          : '%s'", location))

	err := s.generalizeVM(resourceGroupName, computeName)

	if err == nil {
		if isManagedImage {
			s.say(fmt.Sprintf(" -> Image ResourceGroupName   : '%s'", targetManagedImageResourceGroupName))
			s.say(fmt.Sprintf(" -> Image Name                : '%s'", targetManagedImageName))
			s.say(fmt.Sprintf(" -> Image Location            : '%s'", targetManagedImageLocation))
			err = s.captureManagedImage(ctx, targetManagedImageResourceGroupName, targetManagedImageName, imageParameters)
		} else {
			err = s.captureVhd(ctx, resourceGroupName, computeName, vmCaptureParameters)
		}
	}
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

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
