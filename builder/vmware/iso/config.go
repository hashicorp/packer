package iso

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig      `mapstructure:",squash"`
	common.HTTPConfig        `mapstructure:",squash"`
	common.ISOConfig         `mapstructure:",squash"`
	common.FloppyConfig      `mapstructure:",squash"`
	bootcommand.VNCConfig    `mapstructure:",squash"`
	vmwcommon.DriverConfig   `mapstructure:",squash"`
	vmwcommon.HWConfig       `mapstructure:",squash"`
	vmwcommon.OutputConfig   `mapstructure:",squash"`
	vmwcommon.RunConfig      `mapstructure:",squash"`
	vmwcommon.ShutdownConfig `mapstructure:",squash"`
	vmwcommon.SSHConfig      `mapstructure:",squash"`
	vmwcommon.ToolsConfig    `mapstructure:",squash"`
	vmwcommon.VMXConfig      `mapstructure:",squash"`
	vmwcommon.ExportConfig   `mapstructure:",squash"`
	// The size(s) of any additional
    // hard disks for the VM in megabytes. If this is not specified then the VM
    // will only contain a primary hard disk. The builder uses expandable, not
    // fixed-size virtual hard disks, so the actual file representing the disk will
    // not use the full size unless it is full.
	AdditionalDiskSize []uint `mapstructure:"disk_additional_size" required:"false"`
	// The adapter type of the VMware virtual disk
    // to create. This option is for advanced usage, modify only if you know what
    // you're doing. Some of the options you can specify are ide, sata, nvme
    // or scsi (which uses the "lsilogic" scsi interface by default). If you
    // specify another option, Packer will assume that you're specifying a scsi
    // interface of that specified type. For more information, please consult the
    // 
    // Virtual Disk Manager User's Guide for desktop VMware clients.
    // For ESXi, refer to the proper ESXi documentation.
	DiskAdapterType    string `mapstructure:"disk_adapter_type" required:"false"`
	// The filename of the virtual disk that'll be created,
    // without the extension. This defaults to packer.
	DiskName           string `mapstructure:"vmdk_name" required:"false"`
	// The size of the hard disk for the VM in megabytes.
    // The builder uses expandable, not fixed-size virtual hard disks, so the
    // actual file representing the disk will not use the full size unless it
    // is full. By default this is set to 40000 (about 40 GB).
	DiskSize           uint   `mapstructure:"disk_size" required:"false"`
	// The type of VMware virtual disk to create. This
    // option is for advanced usage.
	DiskTypeId         string `mapstructure:"disk_type_id" required:"false"`
	// Either "ovf", "ova" or "vmx", this specifies the output
    // format of the exported virtual machine. This defaults to "ovf".
    // Before using this option, you need to install ovftool. This option
    // currently only works when option remote_type is set to "esx5".
    // Since ovftool is only capable of password based authentication
    // remote_password must be set when exporting the VM.
	Format             string `mapstructure:"format" required:"false"`
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
    // virtual machine.
	GuestOSType string `mapstructure:"guest_os_type" required:"false"`
	// The vmx hardware
    // version
    // for the new virtual machine. Only the default value has been tested, any
    // other value is experimental. Default value is 9.
	Version     string `mapstructure:"version" required:"false"`
	// This is the name of the VMX file for the new virtual
    // machine, without the file extension. By default this is packer-BUILDNAME,
    // where "BUILDNAME" is the name of the build.
	VMName      string `mapstructure:"vm_name" required:"false"`

	VMXDiskTemplatePath string `mapstructure:"vmx_disk_template_path"`
	// Path to a configuration
    // template that defines the
    // contents of the virtual machine VMX file for VMware. This is for advanced
    // users only as this can render the virtual machine non-functional. See
    // below for more information. For basic VMX modifications, try
    // vmx_data first.
	VMXTemplatePath     string `mapstructure:"vmx_template_path" required:"false"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
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
		return nil, nil, err
	}

	// Accumulate any errors and warnings
	var errs *packer.MultiError
	warnings := make([]string, 0)

	isoWarnings, isoErrs := c.ISOConfig.Prepare(&c.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packer.MultiErrorAppend(errs, isoErrs...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HWConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.DriverConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs,
		c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ToolsConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VMXConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VNCConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)

	if c.DiskName == "" {
		c.DiskName = "disk"
	}

	if c.DiskSize == 0 {
		c.DiskSize = 40000
	}

	if c.DiskAdapterType == "" {
		// Default is lsilogic
		c.DiskAdapterType = "lsilogic"
	}

	if !c.SkipCompaction {
		if c.RemoteType == "esx5" {
			if c.DiskTypeId == "" {
				c.SkipCompaction = true
			}
		}
	}

	if c.DiskTypeId == "" {
		// Default is growable virtual disk split in 2GB files.
		c.DiskTypeId = "1"

		if c.RemoteType == "esx5" {
			c.DiskTypeId = "zeroedthick"
		}
	}

	if c.RemoteType == "esx5" {
		if c.DiskTypeId != "thin" && !c.SkipCompaction {
			errs = packer.MultiErrorAppend(
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
			errs = packer.MultiErrorAppend(
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

	// Remote configuration validation
	if c.RemoteType != "" {
		if c.RemoteHost == "" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("remote_host must be specified"))
		}

		if c.RemoteType != "esx5" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Only 'esx5' value is accepted for remote_type"))
		}
	}

	if c.Format != "" {
		if c.RemoteType != "esx5" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("format is only valid when RemoteType=esx5"))
		}
	} else {
		c.Format = "ovf"
	}

	if !(c.Format == "ova" || c.Format == "ovf" || c.Format == "vmx") {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("format must be one of ova, ovf, or vmx"))
	}

	err = c.DriverConfig.Validate(c.SkipExport)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	// Warnings
	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	if c.Headless && c.DisableVNC {
		warnings = append(warnings,
			"Headless mode uses VNC to retrieve output. Since VNC has been disabled,\n"+
				"you won't be able to see any output.")
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
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
