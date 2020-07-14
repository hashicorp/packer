//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type ContentLibraryDestinationConfig
package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/vmware/govmomi/vapi/vcenter"
)

// With this configuration Packer creates a library item in a content library whose content is a virtual machine template created from the just built VM.
// The virtual machine template is stored in a newly created library item.
type ContentLibraryDestinationConfig struct {
	// Name of the library in which the new library item containing the VM template should be created.
	// The Content Library should be of type Local to allow deploying virtual machines.
	Library string `mapstructure:"library"`
	// Name of the library item that will be created. The name of the item should be different from [vm_name](#vm_name).
	// Defaults to [vm_name](#vm_name) + timestamp.
	Name string `mapstructure:"name"`
	// Description of the library item that will be created. Defaults to "Packer imported [vm_name](#vm_name) VM template".
	Description string `mapstructure:"description"`
	// Cluster onto which the virtual machine template should be placed.
	// If cluster and resource_pool are both specified, resource_pool must belong to cluster.
	// If cluster and host are both specified, host must be a member of cluster.
	// Defaults to [cluster](#cluster).
	Cluster string `mapstructure:"cluster"`
	// Virtual machine folder into which the virtual machine template should be placed.
	// Defaults to the same folder as the source virtual machine.
	Folder string `mapstructure:"folder"`
	// Host onto which the virtual machine template should be placed.
	// If host and resource_pool are both specified, resource_pool must belong to host.
	// If host and cluster are both specified, host must be a member of cluster.
	// Defaults to [host](#host).
	Host string `mapstructure:"host"`
	// Resource pool into which the virtual machine template should be placed.
	// Defaults to [resource_pool](#resource_pool). if [resource_pool](#resource_pool) is also unset,
	// the system will attempt to choose a suitable resource pool for the virtual machine template.
	ResourcePool string `mapstructure:"resource_pool"`
	// The datastore for the virtual machine template's configuration and log files.
	// Defaults to the storage backing associated with the library specified by library.
	Datastore string `mapstructure:"datastore"`
	// If set to true, the VM will be destroyed after deploying the template to the Content Library. Defaults to `false`.
	Destroy bool `mapstructure:"destroy"`
}

func (c *ContentLibraryDestinationConfig) Prepare(lc *LocationConfig) []error {
	var errs *packer.MultiError

	if c.Library == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("a library name must be provided"))
	}
	if c.Name == lc.VMName {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("the content library destination name must be different from the VM name"))
	}

	if c.Name == "" {
		// Add timestamp to the the name to differentiate from the original VM
		// otherwise vSphere won't be able to create the template which will be imported
		name, err := interpolate.Render(lc.VMName+"{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
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

	if errs != nil && len(errs.Errors) > 0 {
		return errs.Errors
	}

	return nil
}

type StepImportToContentLibrary struct {
	ContentLibConfig *ContentLibraryDestinationConfig
}

func (s *StepImportToContentLibrary) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

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

	ui.Say(fmt.Sprintf("Importing VM template %s to Content Library...", s.ContentLibConfig.Name))
	err := vm.ImportToContentLibrary(template)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to import VM template %s: %s", s.ContentLibConfig.Name, err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if s.ContentLibConfig.Destroy {
		state.Put("destroy_vm", s.ContentLibConfig.Destroy)
	}

	return multistep.ActionContinue
}

func (s *StepImportToContentLibrary) Cleanup(multistep.StateBag) {
}
