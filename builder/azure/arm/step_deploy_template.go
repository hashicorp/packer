package arm

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepDeployTemplate struct {
	client     *AzureClient
	deploy     func(ctx context.Context, resourceGroupName string, deploymentName string) error
	delete     func(ctx context.Context, client *AzureClient, resourceType string, resourceName string, resourceGroupName string) error
	disk       func(ctx context.Context, resourceGroupName string, computeName string) (string, string, error)
	deleteDisk func(ctx context.Context, imageType string, imageName string, resourceGroupName string) error
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
	step.delete = deleteResource
	step.disk = step.getImageDetails
	step.deleteDisk = step.deleteImage
	return step
}

func (s *StepDeployTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Deploying deployment template ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> DeploymentName    : '%s'", s.name))

	return processStepResult(
		s.deploy(ctx, resourceGroupName, s.name),
		s.error, state)
}

func (s *StepDeployTemplate) Cleanup(state multistep.StateBag) {
	defer s.deleteTemplate(context.Background(), state)

	//Only clean up if this was an existing resource group and the resource group
	//is marked as created
	existingResourceGroup := state.Get(constants.ArmIsExistingResourceGroup).(bool)
	resourceGroupCreated := state.Get(constants.ArmIsResourceGroupCreated).(bool)
	if !existingResourceGroup || !resourceGroupCreated {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say("\nThe resource group was not created by Packer, deleting individual resources ...")

	deploymentName := s.name
	resourceGroupName := state.Get(constants.ArmResourceGroupName).(string)

	// Get image disk details before deleting the image; otherwise we won't be able to
	// delete the disk as the image request will return a 404
	computeName := state.Get(constants.ArmComputeName).(string)
	imageType, imageName, err := s.disk(context.TODO(), resourceGroupName, computeName)

	if err != nil && !strings.Contains(err.Error(), "ResourceNotFound") {
		ui.Error(fmt.Sprintf("Could not retrieve OS Image details: %s", err))
	}

	ui.Say(" -> Deployment Resources within: " + deploymentName)
	if deploymentName != "" {
		maxResources := int32(50)
		deploymentOperations, err := s.client.DeploymentOperationsClient.ListComplete(context.TODO(), resourceGroupName, deploymentName, &maxResources)
		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting resources.  Please delete them manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", resourceGroupName, err))
		}

		for deploymentOperations.NotDone() {
			deploymentOperation := deploymentOperations.Value()
			// Sometimes an empty operation is added to the list by Azure
			if deploymentOperation.Properties.TargetResource == nil {
				if err := deploymentOperations.Next(); err != nil {
					ui.Error(fmt.Sprintf("Error moving to to next deployment operation ...\n\n"+
						"Name: %s\n"+
						"Error: %s", resourceGroupName, err))
					break
				}
				continue
			}

			ui.Say(fmt.Sprintf(" -> %s : '%s'",
				*deploymentOperation.Properties.TargetResource.ResourceType,
				*deploymentOperation.Properties.TargetResource.ResourceName))

			err = s.delete(context.TODO(), s.client,
				*deploymentOperation.Properties.TargetResource.ResourceType,
				*deploymentOperation.Properties.TargetResource.ResourceName,
				resourceGroupName)
			if err != nil {
				ui.Error(fmt.Sprintf("Error deleting resource.  Please delete manually.\n\n"+
					"Name: %s\n"+
					"Error: %s", *deploymentOperation.Properties.TargetResource.ResourceName, err))
			}

			if err = deploymentOperations.Next(); err != nil {
				ui.Error(fmt.Sprintf("Error deleting resources.  Please delete them manually.\n\n"+
					"Name: %s\n"+
					"Error: %s", resourceGroupName, err))
				break
			}
		}

		// The disk is not defined as an operation in the template so it has to be deleted separately
		if imageType == "" && imageName == "" {
			return
		}

		ui.Say(fmt.Sprintf(" -> %s : '%s'", imageType, imageName))
		err = s.deleteDisk(context.TODO(), imageType, imageName, resourceGroupName)
		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting resource.  Please delete manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", imageName, err))
		}
	}
}

func (s *StepDeployTemplate) deployTemplate(ctx context.Context, resourceGroupName string, deploymentName string) error {
	deployment, err := s.factory(s.config)
	if err != nil {
		return err
	}

	f, err := s.client.DeploymentsClient.CreateOrUpdate(ctx, resourceGroupName, deploymentName, *deployment)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, s.client.DeploymentsClient.Client)
	}
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepDeployTemplate) deleteTemplate(ctx context.Context, state multistep.StateBag) error {
	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var deploymentName = s.name
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("Removing the created Deployment object: '%s'", deploymentName))
	f, err := s.client.DeploymentsClient.Delete(ctx, resourceGroupName, deploymentName)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, s.client.DeploymentsClient.Client)
	}

	if err != nil {
		s.say(s.client.LastError.Error())
	}

	return err
}

func (s *StepDeployTemplate) getImageDetails(ctx context.Context, resourceGroupName string, computeName string) (string, string, error) {
	//We can't depend on constants.ArmOSDiskVhd being set
	var imageName, imageType string
	vm, err := s.client.VirtualMachinesClient.Get(ctx, resourceGroupName, computeName, "")
	if err != nil {
		return imageName, imageType, err
	}

	if vm.StorageProfile.OsDisk.Vhd != nil {
		imageType = "image"
		imageName = *vm.StorageProfile.OsDisk.Vhd.URI
	} else {
		imageType = "Microsoft.Compute/disks"
		imageName = *vm.StorageProfile.OsDisk.ManagedDisk.ID
	}

	return imageType, imageName, nil
}

//TODO(paulmey): move to helpers file
func deleteResource(ctx context.Context, client *AzureClient, resourceType string, resourceName string, resourceGroupName string) error {
	switch resourceType {
	case "Microsoft.Compute/virtualMachines":
		f, err := client.VirtualMachinesClient.Delete(ctx, resourceGroupName, resourceName)
		if err == nil {
			err = f.WaitForCompletionRef(ctx, client.VirtualMachinesClient.Client)
		}
		return err
	case "Microsoft.KeyVault/vaults":
		_, err := client.VaultClientDelete.Delete(ctx, resourceGroupName, resourceName)
		return err
	case "Microsoft.Network/networkInterfaces":
		f, err := client.InterfacesClient.Delete(ctx, resourceGroupName, resourceName)
		if err == nil {
			err = f.WaitForCompletionRef(ctx, client.InterfacesClient.Client)
		}
		return err
	case "Microsoft.Network/virtualNetworks":
		f, err := client.VirtualNetworksClient.Delete(ctx, resourceGroupName, resourceName)
		if err == nil {
			err = f.WaitForCompletionRef(ctx, client.VirtualNetworksClient.Client)
		}
		return err
	case "Microsoft.Network/networkSecurityGroups":
		f, err := client.SecurityGroupsClient.Delete(ctx, resourceGroupName, resourceName)
		if err == nil {
			err = f.WaitForCompletionRef(ctx, client.SecurityGroupsClient.Client)
		}
		return err
	case "Microsoft.Network/publicIPAddresses":
		f, err := client.PublicIPAddressesClient.Delete(ctx, resourceGroupName, resourceName)
		if err == nil {
			err = f.WaitForCompletionRef(ctx, client.PublicIPAddressesClient.Client)
		}
		return err
	}
	return nil
}

func (s *StepDeployTemplate) deleteImage(ctx context.Context, imageType string, imageName string, resourceGroupName string) error {
	// Managed disk
	if imageType == "Microsoft.Compute/disks" {
		xs := strings.Split(imageName, "/")
		diskName := xs[len(xs)-1]
		f, err := s.client.DisksClient.Delete(ctx, resourceGroupName, diskName)
		if err == nil {
			err = f.WaitForCompletionRef(ctx, s.client.DisksClient.Client)
		}
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
	if _, err := blob.BreakLease(nil); err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	err = blob.Delete(nil)

	return err
}
