//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type ContentLibraryDestinationConfig
package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/vmware/govmomi/vapi/vcenter"
)

// With this configuration Packer creates a library item in a content library whose content is a VM template
// or an OVF template created from the just built VM.
// The template is stored in a existing or newly created library item.
type ContentLibraryDestinationConfig struct {
	// Name of the library in which the new library item containing the template should be created/updated.
	// The Content Library should be of type Local to allow deploying virtual machines.
	Library string `mapstructure:"library"`
	// Name of the library item that will be created or updated.
	// For VM templates, the name of the item should be different from [vm_name](#vm_name) and
	// the default is [vm_name](#vm_name) + timestamp when not set. VM templates will be always imported to a new library item.
	// For OVF templates, the name defaults to [vm_name](#vm_name) when not set, and if an item with the same name already
	// exists it will be then updated with the new OVF template, otherwise a new item will be created.
	//
	// ~> **Note**: It's not possible to update existing library items with a new VM template. If updating an existing library
	// item is necessary, use an OVF template instead by setting the [ovf](#ovf) option as `true`.
	//
	Name string `mapstructure:"name"`
	// Description of the library item that will be created.
	// This option is not used when importing OVF templates.
	// Defaults to "Packer imported [vm_name](#vm_name) VM template".
	Description string `mapstructure:"description"`
	// Cluster onto which the virtual machine template should be placed.
	// If cluster and resource_pool are both specified, resource_pool must belong to cluster.
	// If cluster and host are both specified, host must be a member of cluster.
	// This option is not used when importing OVF templates.
	// Defaults to [cluster](#cluster).
	Cluster string `mapstructure:"cluster"`
	// Virtual machine folder into which the virtual machine template should be placed.
	// This option is not used when importing OVF templates.
	// Defaults to the same folder as the source virtual machine.
	Folder string `mapstructure:"folder"`
	// Host onto which the virtual machine template should be placed.
	// If host and resource_pool are both specified, resource_pool must belong to host.
	// If host and cluster are both specified, host must be a member of cluster.
	// This option is not used when importing OVF templates.
	// Defaults to [host](#host).
	Host string `mapstructure:"host"`
	// Resource pool into which the virtual machine template should be placed.
	// Defaults to [resource_pool](#resource_pool). if [resource_pool](#resource_pool) is also unset,
	// the system will attempt to choose a suitable resource pool for the virtual machine template.
	ResourcePool string `mapstructure:"resource_pool"`
	// The datastore for the virtual machine template's configuration and log files.
	// This option is not used when importing OVF templates.
	// Defaults to the storage backing associated with the library specified by library.
	Datastore string `mapstructure:"datastore"`
	// If set to true, the VM will be destroyed after deploying the template to the Content Library.
	// Defaults to `false`.
	Destroy bool `mapstructure:"destroy"`
	// When set to true, Packer will import and OVF template to the content library item. Defaults to `false`.
	Ovf bool `mapstructure:"ovf"`
}

func (c *ContentLibraryDestinationConfig) Prepare(lc *LocationConfig) []error {
	var errs *packersdk.MultiError

	if c.Library == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("a library name must be provided"))
	}

	if c.Ovf {
		if c.Name == "" {
			c.Name = lc.VMName
		}
	} else {
		if c.Name == lc.VMName {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("the content library destination name must be different from the VM name"))
		}

		if c.Name == "" {
			// Add timestamp to the name to differentiate from the original VM
			// otherwise vSphere won't be able to create the template which will be imported
			name, err := interpolate.Render(lc.VMName+"{{timestamp}}", nil)
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs,
					fmt.Errorf("unable to parse content library VM template name: %s", err))
			}
			c.Name = name
		}
		if c.Cluster == "" {
			c.Cluster = lc.Cluster
		}
		if c.Host == "" {
			c.Host = lc.Host
		}
		if c.ResourcePool == "" {
			c.ResourcePool = lc.ResourcePool
		}
		if c.Description == "" {
			c.Description = fmt.Sprintf("Packer imported %s VM template", lc.VMName)
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs.Errors
	}

	return nil
}

type StepImportToContentLibrary struct {
	ContentLibConfig *ContentLibraryDestinationConfig
}

func (s *StepImportToContentLibrary) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)
	var err error

	if s.ContentLibConfig.Ovf {
		ui.Say(fmt.Sprintf("Importing VM OVF template %s to Content Library...", s.ContentLibConfig.Name))
		err = s.importOvfTemplate(vm)
	} else {
		ui.Say(fmt.Sprintf("Importing VM template %s to Content Library...", s.ContentLibConfig.Name))
		err = s.importVmTemplate(vm)
	}

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to import template %s: %s", s.ContentLibConfig.Name, err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if s.ContentLibConfig.Destroy {
		state.Put("destroy_vm", s.ContentLibConfig.Destroy)
	}

	return multistep.ActionContinue
}

func (s *StepImportToContentLibrary) importOvfTemplate(vm *driver.VirtualMachineDriver) error {
	ovf := vcenter.OVF{
		Spec: vcenter.CreateSpec{
			Name: s.ContentLibConfig.Name,
		},
		Target: vcenter.LibraryTarget{
			LibraryID: s.ContentLibConfig.Library,
		},
	}
	return vm.ImportOvfToContentLibrary(ovf)
}

func (s *StepImportToContentLibrary) importVmTemplate(vm *driver.VirtualMachineDriver) error {
	template := vcenter.Template{
		Name:        s.ContentLibConfig.Name,
		Description: s.ContentLibConfig.Description,
		Library:     s.ContentLibConfig.Library,
		Placement: &vcenter.Placement{
			Cluster:      s.ContentLibConfig.Cluster,
			Folder:       s.ContentLibConfig.Folder,
			Host:         s.ContentLibConfig.Host,
			ResourcePool: s.ContentLibConfig.ResourcePool,
		},
	}

	if s.ContentLibConfig.Datastore != "" {
		template.VMHomeStorage = &vcenter.DiskStorage{
			Datastore: s.ContentLibConfig.Datastore,
		}
	}

	return vm.ImportToContentLibrary(template)
}

func (s *StepImportToContentLibrary) Cleanup(multistep.StateBag) {
}
