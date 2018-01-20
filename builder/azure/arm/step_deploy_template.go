package arm

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepDeployTemplate struct {
	client     *AzureClient
	deploy     func(resourceGroupName string, deploymentName string, cancelCh <-chan struct{}) error
	delete     func(resourceType string, resourceName string, resourceGroupName string) error
	disk       func(resourceGroupName string, computeName string) (string, string, error)
	deleteDisk func(imageType string, imageName string, resourceGroupName string) error
	say        func(message string)
	error      func(e error)
	config     *Config
	factory    templateFactoryFunc
	name       string
}

func NewStepDeployTemplate(client *AzureClient, ui packer.Ui, config *Config, deploymentName string, factory templateFactoryFunc) *StepDeployTemplate {
	var step = &StepDeployTemplate{
		client:  client,
		say:     func(message string) { ui.Say(message) },
		error:   func(e error) { ui.Error(e.Error()) },
		config:  config,
		factory: factory,
		name:    deploymentName,
	}

	step.deploy = step.deployTemplate
	step.delete = step.deleteOperationResource
	step.disk = step.getImageDetails
	step.deleteDisk = step.deleteImage
	return step
}

func (s *StepDeployTemplate) deployTemplate(resourceGroupName string, deploymentName string, cancelCh <-chan struct{}) error {
	deployment, err := s.factory(s.config)
	if err != nil {
		return err
	}

	_, errChan := s.client.DeploymentsClient.CreateOrUpdate(resourceGroupName, deploymentName, *deployment, cancelCh)

	err = <-errChan
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepDeployTemplate) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Deploying deployment template ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> DeploymentName    : '%s'", s.name))

	result := common.StartInterruptibleTask(
		func() bool { return common.IsStateCancelled(state) },
		func(cancelCh <-chan struct{}) error {
			return s.deploy(resourceGroupName, s.name, cancelCh)
		},
	)

	return processInterruptibleResult(result, s.error, state)
}

func (s *StepDeployTemplate) getImageDetails(resourceGroupName string, computeName string) (string, string, error) {
	//We can't depend on constants.ArmOSDiskVhd being set
	var imageName string
	var imageType string
	vm, err := s.client.VirtualMachinesClient.Get(resourceGroupName, computeName, "")
	if err != nil {
		return imageName, imageType, err
	} else {
		if vm.StorageProfile.OsDisk.Vhd != nil {
			imageType = "image"
			imageName = *vm.StorageProfile.OsDisk.Vhd.URI
		} else {
			imageType = "Microsoft.Compute/disks"
			imageName = *vm.StorageProfile.OsDisk.ManagedDisk.ID
		}
	}
	return imageType, imageName, nil
}

func (s *StepDeployTemplate) deleteOperationResource(resourceType string, resourceName string, resourceGroupName string) error {
	var networkDeleteFunction func(string, string, <-chan struct{}) (<-chan autorest.Response, <-chan error)
	switch resourceType {
	case "Microsoft.Compute/virtualMachines":
		_, errChan := s.client.VirtualMachinesClient.Delete(resourceGroupName,
			resourceName, nil)
		err := <-errChan
		if err != nil {
			return err

		}
	case "Microsoft.KeyVault/vaults":
		_, err := s.client.VaultClientDelete.Delete(resourceGroupName, resourceName)
		return err
	case "Microsoft.Network/networkInterfaces":
		networkDeleteFunction = s.client.InterfacesClient.Delete
	case "Microsoft.Network/virtualNetworks":
		networkDeleteFunction = s.client.VirtualNetworksClient.Delete
	case "Microsoft.Network/publicIPAddresses":
		networkDeleteFunction = s.client.PublicIPAddressesClient.Delete
	}
	if networkDeleteFunction != nil {
		_, errChan := networkDeleteFunction(resourceGroupName, resourceName, nil)
		err := <-errChan
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StepDeployTemplate) deleteImage(imageType string, imageName string, resourceGroupName string) error {
	// Managed disk
	if imageType == "Microsoft.Compute/disks" {
		xs := strings.Split(imageName, "/")
		diskName := xs[len(xs)-1]
		_, errChan := s.client.DisksClient.Delete(resourceGroupName, diskName, nil)
		err := <-errChan
		return err
	}
	// VHD image
	u, err := url.Parse(imageName)
	if err != nil {
		return err
	}
	xs := strings.Split(u.Path, "/")
	if len(xs) < 3 {
		return errors.New("Unable to parse path of image " + imageName)
	}
	var storageAccountName = xs[1]
	var blobName = strings.Join(xs[2:], "/")

	blob := s.client.BlobStorageClient.GetContainerReference(storageAccountName).GetBlobReference(blobName)
	err = blob.Delete(nil)
	return err
}

func (s *StepDeployTemplate) Cleanup(state multistep.StateBag) {
	//Only clean up if this was an existing resource group and the resource group
	//is marked as created
	var existingResourceGroup = state.Get(constants.ArmIsExistingResourceGroup).(bool)
	var resourceGroupCreated = state.Get(constants.ArmIsResourceGroupCreated).(bool)
	if !existingResourceGroup || !resourceGroupCreated {
		return
	}
	ui := state.Get("ui").(packer.Ui)
	ui.Say("\nThe resource group was not created by Packer, deleting individual resources ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var computeName = state.Get(constants.ArmComputeName).(string)
	var deploymentName = s.name
	imageType, imageName, err := s.disk(resourceGroupName, computeName)
	if err != nil {
		ui.Error("Could not retrieve OS Image details")
	}

	ui.Say(" -> Deployment: " + deploymentName)
	if deploymentName != "" {
		maxResources := int32(50)
		deploymentOperations, err := s.client.DeploymentOperationsClient.List(resourceGroupName, deploymentName, &maxResources)
		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting resources.  Please delete them manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", resourceGroupName, err))
		}
		for _, deploymentOperation := range *deploymentOperations.Value {
			// Sometimes an empty operation is added to the list by Azure
			if deploymentOperation.Properties.TargetResource == nil {
				continue
			}
			ui.Say(fmt.Sprintf(" -> %s : '%s'",
				*deploymentOperation.Properties.TargetResource.ResourceType,
				*deploymentOperation.Properties.TargetResource.ResourceName))
			err = s.delete(*deploymentOperation.Properties.TargetResource.ResourceType,
				*deploymentOperation.Properties.TargetResource.ResourceName,
				resourceGroupName)
			if err != nil {
				ui.Error(fmt.Sprintf("Error deleting resource.  Please delete manually.\n\n"+
					"Name: %s\n"+
					"Error: %s", *deploymentOperation.Properties.TargetResource.ResourceName, err))
			}
		}

		// The disk is not defined as an operation in the template so has to be
		// deleted separately
		ui.Say(fmt.Sprintf(" -> %s : '%s'", imageType, imageName))
		err = s.deleteDisk(imageType, imageName, resourceGroupName)
		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting resource.  Please delete manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", imageName, err))
		}
	}
}
