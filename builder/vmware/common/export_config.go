//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type ExportConfig struct {
	// Either "ovf", "ova" or "vmx", this specifies the output
	// format of the exported virtual machine. This defaults to "ovf" for
	// remote (esx) builds, and "vmx" for local builds.
	// Before using this option, you need to install ovftool.
	// Since ovftool is only capable of password based authentication
	// remote_password must be set when exporting the VM from a remote instance.
	// If you are building locally, Packer will create a vmx and then
	// export that vm to an ovf or ova. Packer will not delete the vmx and vmdk
	// files; this is left up to the user if you don't want to keep those
	// files.
	Format string `mapstructure:"format" required:"false"`
	// Extra options to pass to ovftool during export. Each item in the array
	// is a new argument. The options `--noSSLVerify`, `--skipManifestCheck`,
	// and `--targetType` are used by Packer for remote exports, and should not
	// be passed to this argument. For ovf/ova exports from local builds, Packer
	// does not automatically set any ovftool options.
	OVFToolOptions []string `mapstructure:"ovftool_options" required:"false"`
	// Defaults to `false`. When true, Packer will not export the VM. This can
	// be useful if the build output is not the resultant image, but created
	// inside the VM.
	SkipExport bool `mapstructure:"skip_export" required:"false"`
	// Set this to true if you would like to keep a remotely-built
	// VM registered with the remote ESXi server. If you do not need to export
	// the vm, then also set `skip_export: true` in order to avoid unnecessarily
	// using ovftool to export the vm. Defaults to false.
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
