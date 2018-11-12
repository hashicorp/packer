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
	vmwcommon.OutputConfig   `mapstructure:",squash"`
	vmwcommon.RunConfig      `mapstructure:",squash"`
	vmwcommon.ShutdownConfig `mapstructure:",squash"`
	vmwcommon.SSHConfig      `mapstructure:",squash"`
	vmwcommon.ToolsConfig    `mapstructure:",squash"`
	vmwcommon.VMXConfig      `mapstructure:",squash"`
	vmwcommon.ExportConfig   `mapstructure:",squash"`

	// disk drives
	AdditionalDiskSize []uint `mapstructure:"disk_additional_size"`
	DiskAdapterType    string `mapstructure:"disk_adapter_type"`
	DiskName           string `mapstructure:"vmdk_name"`
	DiskSize           uint   `mapstructure:"disk_size"`
	DiskTypeId         string `mapstructure:"disk_type_id"`
	Format             string `mapstructure:"format"`

	// cdrom drive
	CdromAdapterType string `mapstructure:"cdrom_adapter_type"`

	// platform information
	GuestOSType string `mapstructure:"guest_os_type"`
	Version     string `mapstructure:"version"`
	VMName      string `mapstructure:"vm_name"`

	// Network adapter and type
	NetworkAdapterType string `mapstructure:"network_adapter_type"`
	Network            string `mapstructure:"network"`

	// device presence
	Sound bool `mapstructure:"sound"`
	USB   bool `mapstructure:"usb"`

	// communication ports
	Serial   string `mapstructure:"serial"`
	Parallel string `mapstructure:"parallel"`

	VMXDiskTemplatePath string `mapstructure:"vmx_disk_template_path"`
	VMXTemplatePath     string `mapstructure:"vmx_template_path"`

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

	if c.Network == "" {
		c.Network = "nat"
	}

	if !c.Sound {
		c.Sound = false
	}

	if !c.USB {
		c.USB = false
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

	if c.Format == "" {
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
