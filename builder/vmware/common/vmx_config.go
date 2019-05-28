package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type VMXConfig struct {
	// Arbitrary key/values to enter
    // into the virtual machine VMX file. This is for advanced users who want to
    // set properties that aren't yet supported by the builder.
	VMXData           map[string]string `mapstructure:"vmx_data" required:"false"`
	// Identical to vmx_data,
    // except that it is run after the virtual machine is shutdown, and before the
    // virtual machine is exported.
	VMXDataPost       map[string]string `mapstructure:"vmx_data_post" required:"false"`
	// Remove all ethernet interfaces
    // from the VMX file after building. This is for advanced users who understand
    // the ramifications, but is useful for building Vagrant boxes since Vagrant
    // will create ethernet interfaces when provisioning a box. Defaults to
    // false.
	VMXRemoveEthernet bool              `mapstructure:"vmx_remove_ethernet_interfaces" required:"false"`
	// The name that will appear in your vSphere client,
    // and will be used for the vmx basename. This will override the "displayname"
    // value in your vmx file. It will also override the "displayname" if you have
    // set it in the "vmx_data" Packer option. This option is useful if you are
    // chaining vmx builds and want to make sure that the display name of each step
    // in the chain is unique.
	VMXDisplayName    string            `mapstructure:"display_name" required:"false"`
}

func (c *VMXConfig) Prepare(ctx *interpolate.Context) []error {
	return nil
}
