package pvm

import (
	"fmt"
	parallelscommon "github.com/mitchellh/packer/builder/parallels/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig                 `mapstructure:",squash"`
	parallelscommon.FloppyConfig        `mapstructure:",squash"`
	parallelscommon.OutputConfig        `mapstructure:",squash"`
	parallelscommon.RunConfig           `mapstructure:",squash"`
	parallelscommon.SSHConfig           `mapstructure:",squash"`
	parallelscommon.ShutdownConfig      `mapstructure:",squash"`
	parallelscommon.PrlctlConfig        `mapstructure:",squash"`
	parallelscommon.PrlctlVersionConfig `mapstructure:",squash"`

	BootCommand             []string `mapstructure:"boot_command"`
	ParallelsToolsMode      string   `mapstructure:"parallels_tools_mode"`
	ParallelsToolsGuestPath string   `mapstructure:"parallels_tools_guest_path"`
	ParallelsToolsHostPath  string   `mapstructure:"parallels_tools_host_path"`
	SourcePath              string   `mapstructure:"source_path"`
	VMName                  string   `mapstructure:"vm_name"`

	tpl *packer.ConfigTemplate
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	md, err := common.DecodeConfig(c, raws...)
	if err != nil {
		return nil, nil, err
	}

	c.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, nil, err
	}
	c.tpl.UserVars = c.PackerUserVars

	// Defaults
	if c.ParallelsToolsMode == "" {
		c.ParallelsToolsMode = "disable"
	}

	if c.ParallelsToolsGuestPath == "" {
		c.ParallelsToolsGuestPath = "prl-tools.iso"
	}

	if c.ParallelsToolsHostPath == "" {
		c.ParallelsToolsHostPath = "/Applications/Parallels Desktop.app/Contents/Resources/Tools/prl-tools-other.iso"
	}

	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s-{{timestamp}}", c.PackerBuildName)
	}

	// Prepare the errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(c.tpl, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.PrlctlConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.PrlctlVersionConfig.Prepare(c.tpl)...)

	templates := map[string]*string{
		"parallels_tools_mode":       &c.ParallelsToolsMode,
		"parallels_tools_host_paht":  &c.ParallelsToolsHostPath,
		"parallels_tools_guest_path": &c.ParallelsToolsGuestPath,
		"source_path":                &c.SourcePath,
		"vm_name":                    &c.VMName,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = c.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	for i, command := range c.BootCommand {
		if err := c.tpl.Validate(command); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing boot_command[%d]: %s", i, err))
		}
	}

	validMode := false
	validModes := []string{
		parallelscommon.ParallelsToolsModeDisable,
		parallelscommon.ParallelsToolsModeAttach,
		parallelscommon.ParallelsToolsModeUpload,
	}

	for _, mode := range validModes {
		if c.ParallelsToolsMode == mode {
			validMode = true
			break
		}
	}

	if !validMode {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("parallels_tools_mode is invalid. Must be one of: %v", validModes))
	}

	if c.SourcePath == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is required"))
	} else {
		if _, err := os.Stat(c.SourcePath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("source_path is invalid: %s", err))
		}
	}

	// Warnings
	var warnings []string
	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}
