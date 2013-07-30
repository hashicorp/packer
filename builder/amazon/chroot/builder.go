// The chroot package is able to create an Amazon AMI without requiring
// the launch of a new instance for every build. It does this by attaching
// and mounting the root volume of another AMI and chrooting into that
// directory. It then creates an AMI from that attached drive.
package chroot

import (
	"errors"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/builder/common"
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

	AttachedDevicePath string     `mapstructure:"attached_device_path"`
	ChrootMounts       [][]string `mapstructure:"chroot_mounts"`
	CopyFiles          []string   `mapstructure:"copy_files"`
	DevicePath         string     `mapstructure:"device_path"`
	MountCommand       string     `mapstructure:"mount_command"`
	MountPath          string     `mapstructure:"mount_path"`
	SourceAmi          string     `mapstructure:"source_ami"`
	UnmountCommand     string     `mapstructure:"unmount_command"`
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

	if b.config.DevicePath == "" {
		b.config.DevicePath = "/dev/sdh"
	}

	if b.config.AttachedDevicePath == "" {
		b.config.AttachedDevicePath = "/dev/xvdh"
	}

	if b.config.MountCommand == "" {
		b.config.MountCommand = "mount"
	}

	if b.config.MountPath == "" {
		b.config.MountPath = "/var/packer-amazon-chroot/volumes/{{.Device}}"
	}

	if b.config.UnmountCommand == "" {
		b.config.UnmountCommand = "umount"
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare()...)

	if b.config.SourceAmi == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("source_ami is required."))
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
		&StepCreateVolume{},
		&StepAttachVolume{},
		&StepMountDevice{},
		&StepMountExtra{},
		&StepCopyFiles{},
		&StepChrootProvision{},
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
