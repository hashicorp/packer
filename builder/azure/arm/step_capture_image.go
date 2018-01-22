package arm

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCaptureImage struct {
	client              *AzureClient
	generalizeVM        func(resourceGroupName, computeName string) error
	captureVhd          func(resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters, cancelCh <-chan struct{}) error
	captureManagedImage func(resourceGroupName string, computeName string, parameters *compute.Image, cancelCh <-chan struct{}) error
	get                 func(client *AzureClient) *CaptureTemplate
	say                 func(message string)
	error               func(e error)
}

func NewStepCaptureImage(client *AzureClient, ui packer.Ui) *StepCaptureImage {
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
	_, err := s.client.Generalize(resourceGroupName, computeName)
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepCaptureImage) captureImageFromVM(resourceGroupName string, imageName string, image *compute.Image, cancelCh <-chan struct{}) error {
	_, errChan := s.client.ImagesClient.CreateOrUpdate(resourceGroupName, imageName, *image, cancelCh)
	err := <-errChan
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return <-errChan
}

func (s *StepCaptureImage) captureImage(resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters, cancelCh <-chan struct{}) error {
	_, errChan := s.client.Capture(resourceGroupName, computeName, *parameters, cancelCh)
	err := <-errChan
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return <-errChan
}

func (s *StepCaptureImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Capturing image ...")

	var computeName = state.Get(constants.ArmComputeName).(string)
	var location = state.Get(constants.ArmLocation).(string)
	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var vmCaptureParameters = state.Get(constants.ArmVirtualMachineCaptureParameters).(*compute.VirtualMachineCaptureParameters)
	var imageParameters = state.Get(constants.ArmImageParameters).(*compute.Image)

	var isManagedImage = state.Get(constants.ArmIsManagedImage).(bool)
	var targetManagedImageResourceGroupName = state.Get(constants.ArmManagedImageResourceGroupName).(string)
	var targetManagedImageName = state.Get(constants.ArmManagedImageName).(string)
	var targetManagedImageLocation = state.Get(constants.ArmManagedImageLocation).(string)

	s.say(fmt.Sprintf(" -> Compute ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> Compute Name              : '%s'", computeName))
	s.say(fmt.Sprintf(" -> Compute Location          : '%s'", location))

	result := common.StartInterruptibleTask(
		func() bool {
			return common.IsStateCancelled(state)
		},
		func(cancelCh <-chan struct{}) error {
			err := s.generalizeVM(resourceGroupName, computeName)
			if err != nil {
				return err
			}

			if isManagedImage {
				s.say(fmt.Sprintf(" -> Image ResourceGroupName   : '%s'", targetManagedImageResourceGroupName))
				s.say(fmt.Sprintf(" -> Image Name                : '%s'", targetManagedImageName))
				s.say(fmt.Sprintf(" -> Image Location            : '%s'", targetManagedImageLocation))
				return s.captureManagedImage(targetManagedImageResourceGroupName, targetManagedImageName, imageParameters, cancelCh)
			} else {
				return s.captureVhd(resourceGroupName, computeName, vmCaptureParameters, cancelCh)
			}
		})

	// HACK(chrboum): I do not like this.  The capture method should be returning this value
	// instead having to pass in another lambda.  I'm in this pickle because I am using
	// common.StartInterruptibleTask which is not parametric, and only returns a type of error.
	// I could change it to interface{}, but I do not like that solution either.
	//
	// Having to resort to capturing the template via an inspector is hack, and once I can
	// resolve that I can cleanup this code too.  See the comments in azure_client.go for more
	// details.
	template := s.get(s.client)
	state.Put(constants.ArmCaptureTemplate, template)

	return processInterruptibleResult(result, s.error, state)
}

func (*StepCaptureImage) Cleanup(multistep.StateBag) {
}
