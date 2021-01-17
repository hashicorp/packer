//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

// The chroot package is able to create an qcow2 format image without requiring the
// launch of a new instance virtual machine for every build. It does this by attaching and
// mounting the source image and chrooting into that directory.
// It then creates a new qcow2 format image from that attached drive.
package chroot

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/chroot"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// Config is the configuration that is chained through the steps and settable
// from the template.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// Packer builds from a qcow2 format image file, this parameter is the path of
	// source image file
	SourceImage string `mapstructure:"source_image"`
	// This is the path to the directory where the resulting qcow2 image be created.
	// This may be relative or absolute.
	// If relative, the path is relative to the working directory when packer
	// is executed. This directory must not exist prior to running the builder.
	// By default this is output-BUILDNAME where "BUILDNAME" is the
	// name of the build.
	OutputDir string `mapstructure:"output_directory"`
	// The output image file name, ".qcow2" suffix will be added automatically.
	ImageName string `mapstructure:"image_name"`
	// Options to supply the mount command when mounting devices. Each option
	// will be prefixed with -o and supplied to the mount command ran by
	// Packer. Because this command is ran in a shell, user discretion is
	// advised. See this manual page for the mount command for valid file
	// system specific options.
	MountOptions []string `mapstructure:"mount_options" required:"false"`
	// The partition number containing the / partition. By default this is the
	// first partition of the volume, (for example, xvda1) but you can
	// designate the entire block device by setting "mount_partition": "0" in
	// your config, which will mount xvda instead.
	MountPartition string `mapstructure:"mount_partition" required:"false"`
	// The path where the volume will be mounted. This is where the chroot
	// environment will be. This defaults to
	// `/mnt/packer-qemu-chroot-volumes/{{.Device}}`. This is a configuration
	// template where the .Device variable is replaced with the name of the
	// device where the volume is attached.
	MountPath string `mapstructure:"mount_path" required:"false"`
	// This is a list of devices to mount into the chroot environment. This
	// configuration parameter requires some additional documentation which is
	// in the Chroot Mounts section. Please read that section for more
	// information on how to use this.
	ChrootMounts [][]string `mapstructure:"chroot_mounts" required:"false"`
	// How to run shell commands. This defaults to `{{.Command}}`. This may be
	// useful to set if you want to set environmental variables or perhaps run
	// it with sudo or so on. This is a configuration template where the
	// .Command variable is replaced with the command to be run. Defaults to
	// `{{.Command}}`.
	CommandWrapper string `mapstructure:"command_wrapper" required:"false"`
	// Paths to files on the host that will be copied into the
	// chroot environment prior to provisioning. Defaults to /etc/resolv.conf
	// so that DNS lookups work. Pass an empty list to skip copying
	// /etc/resolv.conf. You may need to do this if you're building an image
	// that uses systemd.
	CopyFiles []string `mapstructure:"copy_files" required:"false"`

	ctx interpolate.Context
}

func (c *Config) GetContext() interpolate.Context {
	return c.ctx
}

type wrappedCommandTemplate struct {
	Command string
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"command_wrapper",
				"mount_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Defaults
	if b.config.ChrootMounts == nil {
		b.config.ChrootMounts = make([][]string, 0)
	}

	if len(b.config.ChrootMounts) == 0 {
		b.config.ChrootMounts = [][]string{
			{"proc", "proc", "/proc"},
			{"sysfs", "sysfs", "/sys"},
			{"bind", "/dev", "/dev"},
			{"devpts", "devpts", "/dev/pts"},
			{"binfmt_misc", "binfmt_misc", "/proc/sys/fs/binfmt_misc"},
		}
	}

	// set default copy file if we're not giving our own
	if b.config.CopyFiles == nil {
		b.config.CopyFiles = []string{"/etc/resolv.conf"}
	}

	if b.config.OutputDir == "" {
		b.config.OutputDir = fmt.Sprintf("output-%s", b.config.PackerBuildName)
	}

	if b.config.ImageName == "" {
		b.config.ImageName = fmt.Sprintf("packer-%s.qcow2", b.config.PackerBuildName)
	} else {
		b.config.ImageName += ".qcow2"
	}

	if b.config.CommandWrapper == "" {
		b.config.CommandWrapper = "{{.Command}}"
	}

	if b.config.MountPath == "" {
		b.config.MountPath = "/mnt/packer-qeum-chroot-volumes/{{.Device}}"
	}

	if b.config.MountPartition == "" {
		b.config.MountPartition = "1"
	}

	// Accumulate any errors or warnings
	var errs *packersdk.MultiError
	var warns []string

	if b.config.SourceImage == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("source_image is required."))
	}

	for _, mounts := range b.config.ChrootMounts {
		if len(mounts) != 3 {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("Each chroot_mounts entry should be three elements."))
			break
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warns, errs
	}

	return []string{"Device", "MountPath"}, warns, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("The qemu-chroot builder only works on Linux environments.")
	}

	wrappedCommand := func(command string) (string, error) {
		ictx := b.config.ctx
		ictx.Data = &wrappedCommandTemplate{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &ictx)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", common.CommandWrapper(wrappedCommand))
	generatedData := &packerbuilderdata.GeneratedData{State: state}

	partition, err := strconv.Atoi(b.config.MountPartition)
	if err != nil {
		return nil, errors.New("Invalid MountPartition value.")
	}

	// Build the steps
	steps := []multistep.Step{
		&StepPrepareSourceImage{GeneratedData: generatedData},
		&StepAttachVolume{
			GeneratedData:  generatedData,
			MountPartition: partition,
		},
		&StepMountDevice{
			MountOptions:  b.config.MountOptions,
			GeneratedData: generatedData,
		},
		&chroot.StepMountExtra{
			ChrootMounts: b.config.ChrootMounts,
		},
		&chroot.StepCopyFiles{
			Files: b.config.CopyFiles,
		},
		&chroot.StepChrootProvision{},
		&chroot.StepEarlyCleanup{},
		&StepCompressImage{},
	}

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// Build the artifact and return it
	artifact := &Artifact{
		dir: b.config.OutputDir,
		files: []string{
			state.Get("image_path").(string),
		},
	}

	return artifact, nil
}
