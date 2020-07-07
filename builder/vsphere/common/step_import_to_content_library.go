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

type ContentLibraryDestinationConfig struct {
	Library      string `mapstructure:"library"`
	Name         string `mapstructure:"name"`
	Description  string `mapstructure:"description"`
	Cluster      string `mapstructure:"cluster"`
	Folder       string `mapstructure:"folder"`
	Host         string `mapstructure:"host"`
	ResourcePool string `mapstructure:"resource_pool"`
}

func (c *ContentLibraryDestinationConfig) Prepare(lc *LocationConfig) []error {
	var errs *packer.MultiError

	if c.Library == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("a library name must be provided"))
	}
	if c.Name != "" && c.Name == lc.VMName {
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
	if c.Folder == "" {
		c.Folder = lc.Folder
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

	ui.Say(fmt.Sprintf("Importing VM template %s to Content Library...", s.ContentLibConfig.Name))
	err := vm.ImportToContentLibrary(template)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to import VM template %s: %s", s.ContentLibConfig.Name, err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepImportToContentLibrary) Cleanup(multistep.StateBag) {
}
