//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package ovf

import (
	"fmt"
	"strings"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig             `mapstructure:",squash"`
	common.HTTPConfig               `mapstructure:",squash"`
	common.FloppyConfig             `mapstructure:",squash"`
	bootcommand.BootConfig          `mapstructure:",squash"`
	vboxcommon.ExportConfig         `mapstructure:",squash"`
	vboxcommon.OutputConfig         `mapstructure:",squash"`
	vboxcommon.RunConfig            `mapstructure:",squash"`
	vboxcommon.CommConfig           `mapstructure:",squash"`
	vboxcommon.ShutdownConfig       `mapstructure:",squash"`
	vboxcommon.VBoxManageConfig     `mapstructure:",squash"`
	vboxcommon.VBoxVersionConfig    `mapstructure:",squash"`
	vboxcommon.GuestAdditionsConfig `mapstructure:",squash"`
	// The checksum for the source_path file. The type of the checksum is
	// specified within the checksum field as a prefix, ex: "md5:{$checksum}".
	// The type of the checksum can also be omitted and Packer will try to
	// infer it based on string length. Valid values are "none", "{$checksum}",
	// "md5:{$checksum}", "sha1:{$checksum}", "sha256:{$checksum}",
	// "sha512:{$checksum}" or "file:{$path}". Here is a list of valid checksum
	// values:
	//  * md5:090992ba9fd140077b0661cb75f7ce13
	//  * 090992ba9fd140077b0661cb75f7ce13
	//  * sha1:ebfb681885ddf1234c18094a45bbeafd91467911
	//  * ebfb681885ddf1234c18094a45bbeafd91467911
	//  * sha256:ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
	//  * ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
	//  * file:http://releases.ubuntu.com/20.04/MD5SUMS
	//  * file:file://./local/path/file.sum
	//  * file:./local/path/file.sum
	//  * none
	// Although the checksum will not be verified when it is set to "none",
	// this is not recommended since these files can be very large and
	// corruption does happen from time to time.
	Checksum string `mapstructure:"checksum" required:"true"`
	// The method by which guest additions are
	// made available to the guest for installation. Valid options are upload,
	// attach, or disable. If the mode is attach the guest additions ISO will
	// be attached as a CD device to the virtual machine. If the mode is upload
	// the guest additions ISO will be uploaded to the path specified by
	// guest_additions_path. The default value is upload. If disable is used,
	// guest additions won't be downloaded, either.
	GuestAdditionsMode string `mapstructure:"guest_additions_mode" required:"false"`
	// The path on the guest virtual machine
	// where the VirtualBox guest additions ISO will be uploaded. By default this
	// is VBoxGuestAdditions.iso which should upload into the login directory of
	// the user. This is a configuration
	// template where the Version
	// variable is replaced with the VirtualBox version.
	GuestAdditionsPath string `mapstructure:"guest_additions_path" required:"false"`
	// The interface type to use to mount
	// guest additions when guest_additions_mode is set to attach. Will
	// default to the value set in iso_interface, if iso_interface is set.
	// Will default to "ide", if iso_interface is not set. Options are "ide" and
	// "sata".
	GuestAdditionsInterface string `mapstructure:"guest_additions_interface" required:"false"`
	// The SHA256 checksum of the guest
	// additions ISO that will be uploaded to the guest VM. By default the
	// checksums will be downloaded from the VirtualBox website, so this only needs
	// to be set if you want to be explicit about the checksum.
	GuestAdditionsSHA256 string `mapstructure:"guest_additions_sha256" required:"false"`
	// The URL to the guest additions ISO
	// to upload. This can also be a file URL if the ISO is at a local path. By
	// default, the VirtualBox builder will attempt to find the guest additions ISO
	// on the local file system. If it is not available locally, the builder will
	// download the proper guest additions ISO from the internet.
	GuestAdditionsURL string `mapstructure:"guest_additions_url" required:"false"`
	// Additional flags to pass to
	// VBoxManage import. This can be used to add additional command-line flags
	// such as --eula-accept to accept a EULA in the OVF.
	ImportFlags []string `mapstructure:"import_flags" required:"false"`
	// Additional options to pass to the
	// VBoxManage import. This can be useful for passing keepallmacs or
	// keepnatmacs options for existing ovf images.
	ImportOpts string `mapstructure:"import_opts" required:"false"`
	// The filepath or URL to an OVF or OVA file that acts as the
	// source of this build.
	SourcePath string `mapstructure:"source_path" required:"true"`
	// The path where the OVA should be saved
	// after download. By default, it will go in the packer cache, with a hash of
	// the original filename as its name.
	TargetPath string `mapstructure:"target_path" required:"false"`
	// This is the name of the OVF file for the new virtual
	// machine, without the file extension. By default this is packer-BUILDNAME,
	// where "BUILDNAME" is the name of the build.
	VMName string `mapstructure:"vm_name" required:"false"`
	// Set this to true if you would like to keep
	// the VM registered with virtualbox. Defaults to false.
	KeepRegistered bool `mapstructure:"keep_registered" required:"false"`
	// Defaults to false. When enabled, Packer will
	// not export the VM. Useful if the build output is not the resultant image,
	// but created inside the VM.
	SkipExport bool `mapstructure:"skip_export" required:"false"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"guest_additions_path",
				"guest_additions_url",
				"vboxmanage",
				"vboxmanage_post",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Defaults
	if c.GuestAdditionsMode == "" {
		c.GuestAdditionsMode = "upload"
	}

	if c.GuestAdditionsPath == "" {
		c.GuestAdditionsPath = "VBoxGuestAdditions.iso"
	}
	if c.GuestAdditionsInterface == "" {
		c.GuestAdditionsInterface = "ide"
	}

	if c.VMName == "" {
		c.VMName = fmt.Sprintf(
			"packer-%s-%d", c.PackerBuildName, interpolate.InitTime.Unix())
	}

	// Prepare the errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.CommConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxManageConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxVersionConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.GuestAdditionsConfig.Prepare(&c.ctx)...)

	if c.SourcePath == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is required"))
	}

	validMode := false
	validModes := []string{
		vboxcommon.GuestAdditionsModeDisable,
		vboxcommon.GuestAdditionsModeAttach,
		vboxcommon.GuestAdditionsModeUpload,
	}

	for _, mode := range validModes {
		if c.GuestAdditionsMode == mode {
			validMode = true
			break
		}
	}

	if !validMode {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("guest_additions_mode is invalid. Must be one of: %v", validModes))
	}

	if c.GuestAdditionsSHA256 != "" {
		c.GuestAdditionsSHA256 = strings.ToLower(c.GuestAdditionsSHA256)
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
		return warnings, errs
	}

	// TODO: Write a packer fix and just remove import_opts
	if c.ImportOpts != "" {
		c.ImportFlags = append(c.ImportFlags, "--options", c.ImportOpts)
	}

	return warnings, nil
}
