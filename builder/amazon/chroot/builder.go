// The chroot package is able to create an Amazon AMI without requiring
// the launch of a new instance for every build. It does this by attaching
// and mounting the root volume of another AMI and chrooting into that
// directory. It then creates an AMI from that attached drive.
package chroot

import (
	"errors"
	"log"
	"runtime"

	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/multistep"
)

// The unique ID for this builder
const BuilderId = "mitchellh.amazon.chroot"

// Config is the configuration that is chained through the steps and
// settable from the template.
type Config struct {
	common.PackerConfig       `mapstructure:",squash"`
	awscommon.AMIBlockDevices `mapstructure:",squash"`
	awscommon.AMIConfig       `mapstructure:",squash"`
	awscommon.AccessConfig    `mapstructure:",squash"`

	ChrootMounts      [][]string                 `mapstructure:"chroot_mounts"`
	CommandWrapper    string                     `mapstructure:"command_wrapper"`
	CopyFiles         []string                   `mapstructure:"copy_files"`
	DevicePath        string                     `mapstructure:"device_path"`
	FromScratch       bool                       `mapstructure:"from_scratch"`
	MountOptions      []string                   `mapstructure:"mount_options"`
	MountPartition    int                        `mapstructure:"mount_partition"`
	MountPath         string                     `mapstructure:"mount_path"`
	PostMountCommands []string                   `mapstructure:"post_mount_commands"`
	PreMountCommands  []string                   `mapstructure:"pre_mount_commands"`
	RootDeviceName    string                     `mapstructure:"root_device_name"`
	RootVolumeSize    int64                      `mapstructure:"root_volume_size"`
	SourceAmi         string                     `mapstructure:"source_ami"`
	SourceAmiFilter   awscommon.AmiFilterOptions `mapstructure:"source_ami_filter"`

	ctx interpolate.Context
}

type wrappedCommandTemplate struct {
	Command string
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config.ctx.Funcs = awscommon.TemplateFuncs
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"ami_description",
				"snapshot_tags",
				"tags",
				"command_wrapper",
				"post_mount_commands",
				"pre_mount_commands",
				"mount_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	if b.config.PackerConfig.PackerForce {
		b.config.AMIForceDeregister = true
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
		b.config.CopyFiles = make([]string, 0)
		if !b.config.FromScratch {
			b.config.CopyFiles = []string{"/etc/resolv.conf"}
		}
	}

	if b.config.CommandWrapper == "" {
		b.config.CommandWrapper = "{{.Command}}"
	}

	if b.config.MountPath == "" {
		b.config.MountPath = "/mnt/packer-amazon-chroot-volumes/{{.Device}}"
	}

	if b.config.MountPartition == 0 {
		b.config.MountPartition = 1
	}

	// Accumulate any errors or warnings
	var errs *packer.MultiError
	var warns []string

	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.AMIConfig.Prepare(&b.config.ctx)...)

	for _, mounts := range b.config.ChrootMounts {
		if len(mounts) != 3 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("Each chroot_mounts entry should be three elements."))
			break
		}
	}

	if b.config.FromScratch {
		if b.config.SourceAmi != "" || !b.config.SourceAmiFilter.Empty() {
			warns = append(warns, "source_ami and source_ami_filter are unused when from_scratch is true")
		}
		if b.config.RootVolumeSize == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("root_volume_size is required with from_scratch."))
		}
		if len(b.config.PreMountCommands) == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("pre_mount_commands is required with from_scratch."))
		}
		if b.config.AMIVirtType == "" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("ami_virtualization_type is required with from_scratch."))
		}
		if b.config.RootDeviceName == "" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("root_device_name is required with from_scratch."))
		}
		if len(b.config.AMIMappings) == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("ami_block_device_mappings is required with from_scratch."))
		}
	} else {
		if b.config.SourceAmi == "" && b.config.SourceAmiFilter.Empty() {
			errs = packer.MultiErrorAppend(
				errs, errors.New("source_ami or source_ami_filter is required."))
		}
		if len(b.config.AMIMappings) != 0 {
			warns = append(warns, "ami_block_device_mappings are unused when from_scratch is false")
		}
		if b.config.RootDeviceName != "" {
			warns = append(warns, "root_device_name is unused when from_scratch is false")
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warns, errs
	}

	log.Println(common.ScrubConfig(b.config, b.config.AccessKey, b.config.SecretKey))
	return warns, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("The amazon-chroot builder only works on Linux environments.")
	}

	session, err := b.config.Session()
	if err != nil {
		return nil, err
	}
	ec2conn := ec2.New(session)

	wrappedCommand := func(command string) (string, error) {
		ctx := b.config.ctx
		ctx.Data = &wrappedCommandTemplate{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &ctx)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("ec2", ec2conn)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", CommandWrapper(wrappedCommand))

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepPreValidate{
			DestAmiName:     b.config.AMIName,
			ForceDeregister: b.config.AMIForceDeregister,
		},
		&StepInstanceInfo{},
	}

	if !b.config.FromScratch {
		steps = append(steps,
			&awscommon.StepSourceAMIInfo{
				SourceAmi:          b.config.SourceAmi,
				EnhancedNetworking: b.config.AMIEnhancedNetworking,
				AmiFilters:         b.config.SourceAmiFilter,
			},
			&StepCheckRootDevice{},
		)
	}

	steps = append(steps,
		&StepFlock{},
		&StepPrepareDevice{},
		&StepCreateVolume{
			RootVolumeSize: b.config.RootVolumeSize,
		},
		&StepAttachVolume{},
		&StepEarlyUnflock{},
		&StepPreMountCommands{
			Commands: b.config.PreMountCommands,
		},
		&StepMountDevice{
			MountOptions:   b.config.MountOptions,
			MountPartition: b.config.MountPartition,
		},
		&StepPostMountCommands{
			Commands: b.config.PostMountCommands,
		},
		&StepMountExtra{},
		&StepCopyFiles{},
		&StepChrootProvision{},
		&StepEarlyCleanup{},
		&StepSnapshot{},
		&awscommon.StepDeregisterAMI{
			ForceDeregister:     b.config.AMIForceDeregister,
			ForceDeleteSnapshot: b.config.AMIForceDeleteSnapshot,
			AMIName:             b.config.AMIName,
		},
		&StepRegisterAMI{
			RootVolumeSize: b.config.RootVolumeSize,
		},
		&awscommon.StepCreateEncryptedAMICopy{
			KeyID:             b.config.AMIKmsKeyId,
			EncryptBootVolume: b.config.AMIEncryptBootVolume,
			Name:              b.config.AMIName,
			AMIMappings:       b.config.AMIBlockDevices.AMIMappings,
		},
		&awscommon.StepAMIRegionCopy{
			AccessConfig:      &b.config.AccessConfig,
			Regions:           b.config.AMIRegions,
			RegionKeyIds:      b.config.AMIRegionKMSKeyIDs,
			EncryptBootVolume: b.config.AMIEncryptBootVolume,
			Name:              b.config.AMIName,
		},
		&awscommon.StepModifyAMIAttributes{
			Description:    b.config.AMIDescription,
			Users:          b.config.AMIUsers,
			Groups:         b.config.AMIGroups,
			ProductCodes:   b.config.AMIProductCodes,
			SnapshotUsers:  b.config.SnapshotUsers,
			SnapshotGroups: b.config.SnapshotGroups,
			Ctx:            b.config.ctx,
		},
		&awscommon.StepCreateTags{
			Tags:         b.config.AMITags,
			SnapshotTags: b.config.SnapshotTags,
			Ctx:          b.config.ctx,
		},
	)

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no AMIs, then just return
	if _, ok := state.GetOk("amis"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &awscommon.Artifact{
		Amis:           state.Get("amis").(map[string]string),
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
