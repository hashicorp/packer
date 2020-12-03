package dtl

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepDeployTemplate struct {
	client     *AzureClient
	deploy     func(ctx context.Context, resourceGroupName string, deploymentName string, state multistep.StateBag) error
	delete     func(ctx context.Context, client *AzureClient, resourceType string, resourceName string, resourceGroupName string) error
	disk       func(ctx context.Context, resourceGroupName string, computeName string) (string, string, error)
	deleteDisk func(ctx context.Context, imageType string, imageName string, resourceGroupName string) error
	say        func(message string)
	error      func(e error)
	config     *Config
	factory    templateFactoryFuncDtl
	name       string
}

func NewStepDeployTemplate(client *AzureClient, ui packersdk.Ui, config *Config, deploymentName string, factory templateFactoryFuncDtl) *StepDeployTemplate {
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

func (s *StepDeployTemplate) deployTemplate(ctx context.Context, resourceGroupName string, deploymentName string, state multistep.StateBag) error {

	vmlistPage, err := s.client.DtlVirtualMachineClient.List(ctx, s.config.tmpResourceGroupName, s.config.LabName, "", "", nil, "")

	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	vmList := vmlistPage.Values()
	for i := range vmList {
		if *vmList[i].Name == s.config.tmpComputeName {
			return fmt.Errorf("Error: Virtual Machine %s already exists. Please use another name", s.config.tmpComputeName)
		}
	}

	s.say(fmt.Sprintf("Creating Virtual Machine %s", s.config.tmpComputeName))
	labMachine, err := s.factory(s.config)
	if err != nil {
		return err
	}

	f, err := s.client.DtlLabsClient.CreateEnvironment(ctx, s.config.tmpResourceGroupName, s.config.LabName, *labMachine)

	if err == nil {
		err = f.WaitForCompletionRef(ctx, s.client.DtlLabsClient.Client)
	}
	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}
	expand := "Properties($expand=ComputeVm,Artifacts,NetworkInterface)"

	vm, err := s.client.DtlVirtualMachineClient.Get(ctx, s.config.tmpResourceGroupName, s.config.LabName, s.config.tmpComputeName, expand)
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	xs := strings.Split(*vm.LabVirtualMachineProperties.ComputeID, "/")
	s.config.VMCreationResourceGroup = xs[4]

	s.say(fmt.Sprintf(" -> VM FQDN : '%s'", *vm.Fqdn))

	state.Put(constants.SSHHost, *vm.Fqdn)
	s.config.tmpFQDN = *vm.Fqdn

	// Resuing the Resource group name from common constants as all steps depend on it.
	state.Put(constants.ArmResourceGroupName, s.config.VMCreationResourceGroup)

	s.say(fmt.Sprintf(" -> VM ResourceGroupName : '%s'", s.config.VMCreationResourceGroup))

	return err
}

func (s *StepDeployTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Deploying deployment template ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)

	s.say(fmt.Sprintf(" -> Lab ResourceGroupName : '%s'", resourceGroupName))

	return processStepResult(
		s.deploy(ctx, resourceGroupName, s.name, state),
		s.error, state)
}

func (s *StepDeployTemplate) getImageDetails(ctx context.Context, resourceGroupName string, computeName string) (string, string, error) {
	//We can't depend on constants.ArmOSDiskVhd being set
	var imageName string
	var imageType string
	vm, err := s.client.VirtualMachinesClient.Get(ctx, resourceGroupName, computeName, "")
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

//TODO(paulmey): move to helpers file
func deleteResource(ctx context.Context, client *AzureClient, resourceType string, resourceName string, resourceGroupName string) error {
	switch resourceType {
	case "Microsoft.Compute/virtualMachines":
		f, err := client.VirtualMachinesClient.Delete(ctx, resourceGroupName, resourceName)
		if err == nil {
			err = f.WaitForCompletionRef(ctx, client.VirtualMachinesClient.Client)
		}
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
	err = blob.Delete(nil)
	return err
}

func (s *StepDeployTemplate) Cleanup(state multistep.StateBag) {
	//Only clean up if this was an existing resource group and the resource group
	//is marked as created
	// Just return now
}
