//go:generate mapstructure-to-hcl2 -type Config,RootBlockDevice

// Package bsusurrogate contains a packer.Builder implementation that
// builds a new EBS-backed OMI using an ephemeral instance.
package bsusurrogate

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

const BuilderId = "oapi.outscale.bsusurrogate"

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	osccommon.AccessConfig `mapstructure:",squash"`
	osccommon.RunConfig    `mapstructure:",squash"`
	osccommon.BlockDevices `mapstructure:",squash"`
	osccommon.OMIConfig    `mapstructure:",squash"`

	RootDevice    RootBlockDevice  `mapstructure:"omi_root_device"`
	VolumeRunTags osccommon.TagMap `mapstructure:"run_volume_tags"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {

	b.config.ctx.Funcs = osccommon.TemplateFuncs

	err := config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"omi_description",
				"run_tags",
				"run_volume_tags",
				"snapshot_tags",
				"spot_tags",
				"tags",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	if b.config.PackerConfig.PackerForce {
		b.config.OMIForceDeregister = true
	}

	// Accumulate any errors
	var errs *packersdk.MultiError
	errs = packersdk.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs,
		b.config.OMIConfig.Prepare(&b.config.AccessConfig, &b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.BlockDevices.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.RootDevice.Prepare(&b.config.ctx)...)

	if b.config.OMIVirtType == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("omi_virtualization_type is required."))
	}

	foundRootVolume := false
	for _, launchDevice := range b.config.BlockDevices.LaunchMappings {
		if launchDevice.DeviceName == b.config.RootDevice.SourceDeviceName {
			foundRootVolume = true
		}
	}

	if !foundRootVolume {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("no volume with name '%s' is found", b.config.RootDevice.SourceDeviceName))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packersdk.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return nil, nil, nil

}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (packersdk.Artifact, error) {
	oscConn := b.config.NewOSCClient()

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("osc", oscConn)
	state.Put("accessConfig", &b.config.AccessConfig)
	state.Put("hook", hook)
	state.Put("ui", ui)

	//VMStep

	omiDevices := b.config.BuildOscOMIDevices()
	launchOSCDevices := b.config.BuildOSCLaunchDevices()

	steps := []multistep.Step{
		&osccommon.StepPreValidate{
			DestOmiName:     b.config.OMIName,
			ForceDeregister: b.config.OMIForceDeregister,
		},
		&osccommon.StepSourceOMIInfo{
			SourceOmi:   b.config.SourceOmi,
			OmiFilters:  b.config.SourceOmiFilter,
			OMIVirtType: b.config.OMIVirtType, //TODO: Remove if it is not used
		},
		&osccommon.StepNetworkInfo{
			NetId:               b.config.NetId,
			NetFilter:           b.config.NetFilter,
			SecurityGroupIds:    b.config.SecurityGroupIds,
			SecurityGroupFilter: b.config.SecurityGroupFilter,
			SubnetId:            b.config.SubnetId,
			SubnetFilter:        b.config.SubnetFilter,
			SubregionName:       b.config.Subregion,
		},
		&osccommon.StepKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.RunConfig.Comm,
			DebugKeyPath: fmt.Sprintf("oapi_%s", b.config.PackerBuildName),
		},
		&osccommon.StepPublicIp{
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			Debug:                    b.config.PackerDebug,
		},
		&osccommon.StepSecurityGroup{
			SecurityGroupFilter:   b.config.SecurityGroupFilter,
			SecurityGroupIds:      b.config.SecurityGroupIds,
			CommConfig:            &b.config.RunConfig.Comm,
			TemporarySGSourceCidr: b.config.TemporarySGSourceCidr,
		},
		&osccommon.StepCleanupVolumes{
			BlockDevices: b.config.BlockDevices,
		},
		&osccommon.StepRunSourceVm{
			BlockDevices:                b.config.BlockDevices,
			Comm:                        &b.config.RunConfig.Comm,
			Ctx:                         b.config.ctx,
			Debug:                       b.config.PackerDebug,
			BsuOptimized:                b.config.BsuOptimized,
			EnableT2Unlimited:           b.config.EnableT2Unlimited,
			ExpectedRootDevice:          osccommon.RunSourceVmBSUExpectedRootDevice,
			IamVmProfile:                b.config.IamVmProfile,
			VmInitiatedShutdownBehavior: b.config.VmInitiatedShutdownBehavior,
			VmType:                      b.config.VmType,
			IsRestricted:                false,
			SourceOMI:                   b.config.SourceOmi,
			Tags:                        b.config.RunTags,
			UserData:                    b.config.UserData,
			UserDataFile:                b.config.UserDataFile,
			VolumeTags:                  b.config.VolumeRunTags,
		},
		&osccommon.StepGetPassword{
			Debug:     b.config.PackerDebug,
			Comm:      &b.config.RunConfig.Comm,
			Timeout:   b.config.WindowsPasswordTimeout,
			BuildName: b.config.PackerBuildName,
		},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: osccommon.OscSSHHost(
				oscConn.VmApi,
				b.config.SSHInterface),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.RunConfig.Comm,
		},
		&osccommon.StepStopBSUBackedVm{
			Skip:          false,
			DisableStopVm: b.config.DisableStopVm,
		},
		&StepSnapshotVolumes{
			LaunchDevices: launchOSCDevices,
		},
		&osccommon.StepDeregisterOMI{
			AccessConfig:        &b.config.AccessConfig,
			ForceDeregister:     b.config.OMIForceDeregister,
			ForceDeleteSnapshot: b.config.OMIForceDeleteSnapshot,
			OMIName:             b.config.OMIName,
			Regions:             b.config.OMIRegions,
		},
		&StepRegisterOMI{
			RootDevice:    b.config.RootDevice,
			OMIDevices:    omiDevices,
			LaunchDevices: launchOSCDevices,
			RawRegion:     b.config.RawRegion,
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
	}

	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	//Build the artifact
	if omis, ok := state.GetOk("omis"); ok {
		// Build the artifact and return it
		artifact := &osccommon.Artifact{
			Omis:           omis.(map[string]string),
			BuilderIdValue: BuilderId,
			StateData:      map[string]interface{}{"generated_data": state.Get("generated_data")},
		}

		return artifact, nil
	}

	return nil, nil
}
