// The chroot package is able to create an Amazon AMI without requiring
// the launch of a new instance for every build. It does this by attaching
// and mounting the root volume of another AMI and chrooting into that
// directory. It then creates an AMI from that attached drive.
package chroot

import (
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"runtime"
)

// The unique ID for this builder
const BuilderId = "mitchellh.amazon.chroot"

// Config is the configuration that is chained through the steps and
// settable from the template.
type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`
	awscommon.AMIConfig    `mapstructure:",squash"`

	ChrootMounts   [][]string `mapstructure:"chroot_mounts"`
	CopyFiles      []string   `mapstructure:"copy_files"`
	DevicePath     string     `mapstructure:"device_path"`
	MountCommand   string     `mapstructure:"mount_command"`
	MountPath      string     `mapstructure:"mount_path"`
	SourceAmi      string     `mapstructure:"source_ami"`
	UnmountCommand string     `mapstructure:"unmount_command"`

	tpl *packer.ConfigTemplate
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Defaults
	if b.config.ChrootMounts == nil {
		b.config.ChrootMounts = make([][]string, 0)
	}

	if b.config.CopyFiles == nil {
		b.config.CopyFiles = make([]string, 0)
	}

	if len(b.config.ChrootMounts) == 0 {
		b.config.ChrootMounts = [][]string{
			[]string{"proc", "proc", "/proc"},
			[]string{"sysfs", "sysfs", "/sys"},
			[]string{"bind", "/dev", "/dev"},
			[]string{"devpts", "devpts", "/dev/pts"},
			[]string{"binfmt_misc", "binfmt_misc", "/proc/sys/fs/binfmt_misc"},
		}
	}

	if len(b.config.CopyFiles) == 0 {
		b.config.CopyFiles = []string{"/etc/resolv.conf"}
	}

	if b.config.MountCommand == "" {
		b.config.MountCommand = "mount"
	}

	if b.config.MountPath == "" {
		b.config.MountPath = "packer-amazon-chroot-volumes/{{.Device}}"
	}

	if b.config.UnmountCommand == "" {
		b.config.UnmountCommand = "umount"
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.AMIConfig.Prepare(b.config.tpl)...)

	for i, mounts := range b.config.ChrootMounts {
		if len(mounts) != 3 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("Each chroot_mounts entry should be three elements."))
			break
		}

		for j, entry := range mounts {
			b.config.ChrootMounts[i][j], err = b.config.tpl.Process(entry, nil)
			if err != nil {
				errs = packer.MultiErrorAppend(errs,
					fmt.Errorf("Error processing chroot_mounts[%d][%d]: %s",
						i, j, err))
			}
		}
	}

	for i, file := range b.config.CopyFiles {
		var err error
		b.config.CopyFiles[i], err = b.config.tpl.Process(file, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing copy_files[%d]: %s",
					i, err))
		}
	}

	if b.config.SourceAmi == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("source_ami is required."))
	}

	templates := map[string]*string{
		"device_path":     &b.config.DevicePath,
		"mount_command":   &b.config.MountCommand,
		"source_ami":      &b.config.SourceAmi,
		"unmount_command": &b.config.UnmountCommand,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	log.Printf("Config: %+v", b.config)
	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("The amazon-chroot builder only works on Linux environments.")
	}

	region, err := b.config.Region()
	if err != nil {
		return nil, err
	}

	auth, err := b.config.AccessConfig.Auth()
	if err != nil {
		return nil, err
	}

	ec2conn := ec2.New(auth, region)

	// Setup the state bag and initial state for the steps
	state := make(map[string]interface{})
	state["config"] = &b.config
	state["ec2"] = ec2conn
	state["hook"] = hook
	state["ui"] = ui

	// Build the steps
	steps := []multistep.Step{
		&StepInstanceInfo{},
		&StepSourceAMIInfo{},
		&StepFlock{},
		&StepPrepareDevice{},
		&StepCreateVolume{},
		&StepAttachVolume{},
		&StepEarlyUnflock{},
		&StepMountDevice{},
		&StepMountExtra{},
		&StepCopyFiles{},
		&StepChrootProvision{},
		&StepEarlyCleanup{},
		&StepSnapshot{},
		&StepRegisterAMI{},
		&awscommon.StepModifyAMIAttributes{
			Description: b.config.AMIDescription,
			Users:       b.config.AMIUsers,
			Groups:      b.config.AMIGroups,
		},
		&awscommon.StepAMIRegionCopy{
			Regions: b.config.AMIRegions,
			Tags:    b.config.AMITags,
		},
		&awscommon.StepCreateTags{
			Tags: b.config.AMITags,
		},
	}

	// Run!
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state["error"]; ok {
		return nil, rawErr.(error)
	}

	// If there are no AMIs, then just return
	if _, ok := state["amis"]; !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &awscommon.Artifact{
		Amis:           state["amis"].(map[string]string),
		BuilderIdValue: BuilderId,
		Conn:           ec2conn,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
