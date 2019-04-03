// The chroot package is able to create an Amazon AMI without requiring
// the launch of a new instance for every build. It does this by attaching
// and mounting the root volume of another AMI and chrooting into that
// directory. It then creates an AMI from that attached drive.
package chroot

import (
	"context"
	"errors"
	"runtime"

	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
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
	NVMEDevicePath    string                     `mapstructure:"nvme_device_path"`
	FromScratch       bool                       `mapstructure:"from_scratch"`
	MountOptions      []string                   `mapstructure:"mount_options"`
	MountPartition    string                     `mapstructure:"mount_partition"`
	MountPath         string                     `mapstructure:"mount_path"`
	PostMountCommands []string                   `mapstructure:"post_mount_commands"`
	PreMountCommands  []string                   `mapstructure:"pre_mount_commands"`
	RootDeviceName    string                     `mapstructure:"root_device_name"`
	RootVolumeSize    int64                      `mapstructure:"root_volume_size"`
	RootVolumeType    string                     `mapstructure:"root_volume_type"`
	SourceAmi         string                     `mapstructure:"source_ami"`
	SourceAmiFilter   awscommon.AmiFilterOptions `mapstructure:"source_ami_filter"`
	RootVolumeTags    awscommon.TagMap           `mapstructure:"root_volume_tags"`

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
				"root_volume_tags",
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

	if b.config.MountPartition == "" {
		b.config.MountPartition = "1"
	}

	// Accumulate any errors or warnings
	var errs *packer.MultiError
	var warns []string

	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs,
		b.config.AMIConfig.Prepare(&b.config.AccessConfig, &b.config.ctx)...)

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
		if len(b.config.AMIMappings) > 0 && b.config.RootDeviceName != "" {
			if b.config.RootVolumeSize == 0 {
				// Although, they can specify the device size in the block device mapping, it's easier to
				// be specific here.
				errs = packer.MultiErrorAppend(
					errs, errors.New("root_volume_size is required if ami_block_device_mappings is specified"))
			}
			warns = append(warns, "ami_block_device_mappings from source image will be completely overwritten")
		} else if len(b.config.AMIMappings) > 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("If ami_block_device_mappings is specified, root_device_name must be specified"))
		} else if b.config.RootDeviceName != "" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("If root_device_name is specified, ami_block_device_mappings must be specified"))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warns, errs
	}

	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return warns, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
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
	state.Put("access_config", &b.config.AccessConfig)
	state.Put("ami_config", &b.config.AMIConfig)
	state.Put("ec2", ec2conn)
	state.Put("awsSession", session)
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
				SourceAmi:                b.config.SourceAmi,
				EnableAMISriovNetSupport: b.config.AMISriovNetSupport,
				EnableAMIENASupport:      b.config.AMIENASupport,
				AmiFilters:               b.config.SourceAmiFilter,
				AMIVirtType:              b.config.AMIVirtType,
			},
			&StepCheckRootDevice{},
		)
	}

	steps = append(steps,
		&StepFlock{},
		&StepPrepareDevice{},
		&StepCreateVolume{
			RootVolumeType: b.config.RootVolumeType,
			RootVolumeSize: b.config.RootVolumeSize,
			RootVolumeTags: b.config.RootVolumeTags,
			Ctx:            b.config.ctx,
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
			AccessConfig:        &b.config.AccessConfig,
			ForceDeregister:     b.config.AMIForceDeregister,
			ForceDeleteSnapshot: b.config.AMIForceDeleteSnapshot,
			AMIName:             b.config.AMIName,
			Regions:             b.config.AMIRegions,
		},
		&StepRegisterAMI{
			RootVolumeSize:           b.config.RootVolumeSize,
			EnableAMISriovNetSupport: b.config.AMISriovNetSupport,
			EnableAMIENASupport:      b.config.AMIENASupport,
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
	b.runner.Run(ctx, state)

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
		Session:        session,
	}

	return artifact, nil
}
