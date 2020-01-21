//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

type ExportConfig struct {
	// Either "ovf", "ova" or "vmx", this specifies the output
	// format of the exported virtual machine. This defaults to "ovf".
	// Before using this option, you need to install ovftool. This option
	// currently only works when option remote_type is set to "esx5".
	// Since ovftool is only capable of password based authentication
	// remote_password must be set when exporting the VM.
	Format string `mapstructure:"format" required:"false"`
	// Extra options to pass to ovftool during export. Each item in the array
	// is a new argument. The options `--noSSLVerify`, `--skipManifestCheck`,
	// and `--targetType` are reserved, and should not be passed to this
	// argument. Currently, exporting the build VM (with ovftool) is only
	// supported when building on ESXi e.g. when `remote_type` is set to
	// `esx5`. See the [Building on a Remote vSphere
	// Hypervisor](/docs/builders/vmware-iso.html#building-on-a-remote-vsphere-hypervisor)
	// section below for more info.
	OVFToolOptions []string `mapstructure:"ovftool_options" required:"false"`
	// Defaults to `false`. When enabled, Packer will not export the VM. Useful
	// if the build output is not the resultant image, but created inside the
	// VM. Currently, exporting the build VM is only supported when building on
	// ESXi e.g. when `remote_type` is set to `esx5`. See the [Building on a
	// Remote vSphere
	// Hypervisor](/docs/builders/vmware-iso.html#building-on-a-remote-vsphere-hypervisor)
	// section below for more info.
	SkipExport bool `mapstructure:"skip_export" required:"false"`
	// Set this to true if you would like to keep
	// the VM registered with the remote ESXi server. If you do not need to export
	// the vm, then also set skip_export: true in order to avoid an unnecessary
	// step of using ovftool to export the vm. Defaults to false.
	KeepRegistered bool `mapstructure:"keep_registered" required:"false"`
	// VMware-created disks are defragmented and
	// compacted at the end of the build process using vmware-vdiskmanager or
	// vmkfstools in ESXi. In certain rare cases, this might actually end up
	// making the resulting disks slightly larger. If you find this to be the case,
	// you can disable compaction using this configuration value. Defaults to
	// false. Default to true for ESXi when disk_type_id is not explicitly
	// defined and false otherwise.
	SkipCompaction bool `mapstructure:"skip_compaction" required:"false"`
}

func (c *ExportConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if c.Format != "" {
		if !(c.Format == "ova" || c.Format == "ovf" || c.Format == "vmx") {
			errs = append(
				errs, fmt.Errorf("format must be one of ova, ovf, or vmx"))
		}
	}
	return errs
}
