//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package iso

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/shutdowncommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig            `mapstructure:",squash"`
	commonsteps.HTTPConfig         `mapstructure:",squash"`
	commonsteps.ISOConfig          `mapstructure:",squash"`
	commonsteps.FloppyConfig       `mapstructure:",squash"`
	commonsteps.CDConfig           `mapstructure:",squash"`
	bootcommand.VNCConfig          `mapstructure:",squash"`
	vmwcommon.DriverConfig         `mapstructure:",squash"`
	vmwcommon.HWConfig             `mapstructure:",squash"`
	vmwcommon.OutputConfig         `mapstructure:",squash"`
	vmwcommon.RunConfig            `mapstructure:",squash"`
	shutdowncommand.ShutdownConfig `mapstructure:",squash"`
	vmwcommon.SSHConfig            `mapstructure:",squash"`
	vmwcommon.ToolsConfig          `mapstructure:",squash"`
	vmwcommon.VMXConfig            `mapstructure:",squash"`
	vmwcommon.ExportConfig         `mapstructure:",squash"`
	vmwcommon.DiskConfig           `mapstructure:",squash"`
	// The size of the hard disk for the VM in megabytes.
	// The builder uses expandable, not fixed-size virtual hard disks, so the
	// actual file representing the disk will not use the full size unless it
	// is full. By default this is set to 40000 (about 40 GB).
	DiskSize uint `mapstructure:"disk_size" required:"false"`
	// The adapter type (or bus) that will be used
	// by the cdrom device. This is chosen by default based on the disk adapter
	// type. VMware tends to lean towards ide for the cdrom device unless
	// sata is chosen for the disk adapter and so Packer attempts to mirror
	// this logic. This field can be specified as either ide, sata, or scsi.
	CdromAdapterType string `mapstructure:"cdrom_adapter_type" required:"false"`
	// The guest OS type being installed. This will be
	// set in the VMware VMX. By default this is other. By specifying a more
	// specific OS type, VMware may perform some optimizations or virtual hardware
	// changes to better support the operating system running in the
	// virtual machine. Valid values differ by platform and version numbers, and may
	// not match other VMware API's representation of the guest OS names. Consult your
	// platform for valid values.
	GuestOSType string `mapstructure:"guest_os_type" required:"false"`
	// The [vmx hardware
	// version](http://kb.vmware.com/selfservice/microsites/search.do?language=en_US&cmd=displayKC&externalId=1003746)
	// for the new virtual machine. Only the default value has been tested, any
	// other value is experimental. Default value is `9`.
	Version string `mapstructure:"version" required:"false"`
	// This is the name of the VMX file for the new virtual
	// machine, without the file extension. By default this is packer-BUILDNAME,
	// where "BUILDNAME" is the name of the build.
	VMName string `mapstructure:"vm_name" required:"false"`

	VMXDiskTemplatePath string `mapstructure:"vmx_disk_template_path"`
	// Path to a [configuration template](/docs/templates/engine) that
	// defines the contents of the virtual machine VMX file for VMware. The
	// engine has access to the template variables `{{ .DiskNumber }}` and
	// `{{ .DiskName }}`.
	//
	// This is for **advanced users only** as this can render the virtual machine
	// non-functional. See below for more information. For basic VMX
	// modifications, try `vmx_data` first.
	VMXTemplatePath string `mapstructure:"vmx_template_path" required:"false"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"tools_upload_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors and warnings
	var warnings []string
	var errs *packersdk.MultiError

	runConfigWarnings, runConfigErrs := c.RunConfig.Prepare(&c.ctx, &c.DriverConfig)
	warnings = append(warnings, runConfigWarnings...)
	errs = packersdk.MultiErrorAppend(errs, runConfigErrs...)
	isoWarnings, isoErrs := c.ISOConfig.Prepare(&c.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packersdk.MultiErrorAppend(errs, isoErrs...)
	errs = packersdk.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.HWConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packersdk.MultiErrorAppend(errs, c.DriverConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.ToolsConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.CDConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.VNCConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.VMXConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.DiskConfig.Prepare(&c.ctx)...)

	if c.DiskSize == 0 {
		c.DiskSize = 40000
	}

	if c.DiskTypeId == "" {
		// Default is growable virtual disk split in 2GB files.
		c.DiskTypeId = "1"

		if c.RemoteType == "esx5" {
			c.DiskTypeId = "zeroedthick"
			c.SkipCompaction = true
		}
	}

	if c.RemoteType == "esx5" {
		if c.DiskTypeId != "thin" && !c.SkipCompaction {
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("skip_compaction must be 'true' for disk_type_id: %s", c.DiskTypeId))
		}
	}

	if c.GuestOSType == "" {
		c.GuestOSType = "other"
	}

	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s", c.PackerBuildName)
	}

	if c.Version == "" {
		c.Version = "9"
	}

	if c.VMXTemplatePath != "" {
		if err := c.validateVMXTemplatePath(); err != nil {
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("vmx_template_path is invalid: %s", err))
		}
	} else {
		warn := c.checkForVMXTemplateAndVMXDataCollisions()
		if warn != "" {
			warnings = append(warnings, warn)
		}
	}

	if c.HWConfig.Network == "" {
		c.HWConfig.Network = "nat"
	}

	if c.Format == "" {
		if c.RemoteType == "" {
			c.Format = "vmx"
		} else {
			c.Format = "ovf"
		}
	}

	if c.RemoteType == "" {
		if c.Format == "vmx" {
			// if we're building locally and want a vmx, there's nothing to export.
			// Set skip export flag here to keep the export step from attempting
			// an unneded export
			c.SkipExport = true
		}
		if c.Headless && c.DisableVNC {
			warnings = append(warnings,
				"Headless mode uses VNC to retrieve output. Since VNC has been disabled,\n"+
					"you won't be able to see any output.")
		}
	}

	err = c.DriverConfig.Validate(c.SkipExport)
	if err != nil {
		errs = packersdk.MultiErrorAppend(errs, err)
	}

	if c.CdromAdapterType != "" {
		c.CdromAdapterType = strings.ToLower(c.CdromAdapterType)
		if c.CdromAdapterType != "ide" && c.CdromAdapterType != "sata" && c.CdromAdapterType != "scsi" {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("cdrom_adapter_type must be one of ide, sata, or scsi"))
		}
	}

	// Warnings
	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

func (c *Config) checkForVMXTemplateAndVMXDataCollisions() string {
	if c.VMXTemplatePath != "" {
		return ""
	}

	var overridden []string
	tplLines := strings.Split(DefaultVMXTemplate, "\n")
	tplLines = append(tplLines,
		fmt.Sprintf("%s0:0.present", strings.ToLower(c.DiskAdapterType)),
		fmt.Sprintf("%s0:0.fileName", strings.ToLower(c.DiskAdapterType)),
		fmt.Sprintf("%s0:0.deviceType", strings.ToLower(c.DiskAdapterType)),
		fmt.Sprintf("%s0:1.present", strings.ToLower(c.DiskAdapterType)),
		fmt.Sprintf("%s0:1.fileName", strings.ToLower(c.DiskAdapterType)),
		fmt.Sprintf("%s0:1.deviceType", strings.ToLower(c.DiskAdapterType)),
	)

	for _, line := range tplLines {
		if strings.Contains(line, `{{`) {
			key := line[:strings.Index(line, " =")]
			if _, ok := c.VMXData[key]; ok {
				overridden = append(overridden, key)
			}
		}
	}

	if len(overridden) > 0 {
		warnings := fmt.Sprintf("Your vmx data contains the following "+
			"variable(s), which Packer normally sets when it generates its "+
			"own default vmx template. This may cause your build to fail or "+
			"behave unpredictably: %s", strings.Join(overridden, ", "))
		return warnings
	}
	return ""
}

// Make sure custom vmx template exists and that data can be read from it
func (c *Config) validateVMXTemplatePath() error {
	f, err := os.Open(c.VMXTemplatePath)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	return interpolate.Validate(string(data), &c.ctx)
}
