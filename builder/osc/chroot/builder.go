//go:generate mapstructure-to-hcl2 -type Config

// Package chroot is able to create an Outscale OMI without requiring
// the launch of a new instance for every build. It does this by attaching
// and mounting the root volume of another OMI and chrooting into that
// directory. It then creates an OMI from that attached drive.
package chroot

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"runtime"

	"github.com/hashicorp/hcl/v2/hcldec"
	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
)

// The unique ID for this builder
const BuilderId = "oapi.outscale.chroot"

// Config is the configuration that is chained through the steps and
// settable from the template.
type Config struct {
	common.PackerConfig       `mapstructure:",squash"`
	osccommon.OMIBlockDevices `mapstructure:",squash"`
	osccommon.OMIConfig       `mapstructure:",squash"`
	osccommon.AccessConfig    `mapstructure:",squash"`

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
	SourceOMI         string                     `mapstructure:"source_omi"`
	SourceOMIFilter   osccommon.OmiFilterOptions `mapstructure:"source_omi_filter"`
	RootVolumeTags    osccommon.TagMap           `mapstructure:"root_volume_tags"`

	ctx interpolate.Context
}

type wrappedCommandTemplate struct {
	Command string
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config.ctx.Funcs = osccommon.TemplateFuncs
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"omi_description",
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
		b.config.OMIForceDeregister = true
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
		b.config.MountPath = "/mnt/packer-outscale-chroot-volumes/{{.Device}}"
	}

	if b.config.MountPartition == "" {
		b.config.MountPartition = "1"
	}

	// Accumulate any errors or warnings
	var errs *packer.MultiError
	var warns []string

	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs,
		b.config.OMIConfig.Prepare(&b.config.AccessConfig, &b.config.ctx)...)

	for _, mounts := range b.config.ChrootMounts {
		if len(mounts) != 3 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("Each chroot_mounts entry should be three elements."))
			break
		}
	}

	if b.config.FromScratch {
		if b.config.SourceOMI != "" || !b.config.SourceOMIFilter.Empty() {
			warns = append(warns, "source_omi and source_omi_filter are unused when from_scratch is true")
		}
		if b.config.RootVolumeSize == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("root_volume_size is required with from_scratch."))
		}
		if len(b.config.PreMountCommands) == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("pre_mount_commands is required with from_scratch."))
		}
		if b.config.OMIVirtType == "" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("omi_virtualization_type is required with from_scratch."))
		}
		if b.config.RootDeviceName == "" {
			errs = packer.MultiErrorAppend(
				errs, errors.New("root_device_name is required with from_scratch."))
		}
		if len(b.config.OMIMappings) == 0 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("omi_block_device_mappings is required with from_scratch."))
		}
	} else {
		if b.config.SourceOMI == "" && b.config.SourceOMIFilter.Empty() {
			errs = packer.MultiErrorAppend(
				errs, errors.New("source_omi or source_omi_filter is required."))
		}
		if len(b.config.OMIMappings) != 0 {
			warns = append(warns, "omi_block_device_mappings are unused when from_scratch is false")
		}
		if b.config.RootDeviceName != "" {
			warns = append(warns, "root_device_name is unused when from_scratch is false")
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
		return nil, errors.New("The outscale-chroot builder only works on Linux environments.")
	}

	clientConfig, err := b.config.Config()
	if err != nil {
		return nil, err
	}

	skipClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	oapiconn := oapi.NewClient(clientConfig, skipClient)

	wrappedCommand := func(command string) (string, error) {
		ctx := b.config.ctx
		ctx.Data = &wrappedCommandTemplate{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &ctx)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("oapi", oapiconn)
	state.Put("clientConfig", clientConfig)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", CommandWrapper(wrappedCommand))

	// Build the steps
	steps := []multistep.Step{
		&osccommon.StepPreValidate{
			DestOmiName:     b.config.OMIName,
			ForceDeregister: b.config.OMIForceDeregister,
		},
		&StepVmInfo{},
	}

	if !b.config.FromScratch {
		steps = append(steps,
			&osccommon.StepSourceOMIInfo{
				SourceOmi:   b.config.SourceOMI,
				OmiFilters:  b.config.SourceOMIFilter,
				OMIVirtType: b.config.OMIVirtType,
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
		&StepLinkVolume{},
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
		&osccommon.StepDeregisterOMI{
			AccessConfig:        &b.config.AccessConfig,
			ForceDeregister:     b.config.OMIForceDeregister,
			ForceDeleteSnapshot: b.config.OMIForceDeleteSnapshot,
			OMIName:             b.config.OMIName,
			Regions:             b.config.OMIRegions,
		},
		&StepCreateOMI{
			RootVolumeSize: b.config.RootVolumeSize,
		},
		&osccommon.StepUpdateOMIAttributes{
			AccountIds:         b.config.OMIAccountIDs,
			SnapshotAccountIds: b.config.SnapshotAccountIDs,
			Ctx:                b.config.ctx,
		},
		&osccommon.StepCreateTags{
			Tags:         b.config.OMITags,
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

	// If there are no OMIs, then just return
	if _, ok := state.GetOk("omis"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &osccommon.Artifact{
		Omis:           state.Get("omis").(map[string]string),
		BuilderIdValue: BuilderId,
		Config:         clientConfig,
	}

	return artifact, nil
}
