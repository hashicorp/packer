//go:generate mapstructure-to-hcl2 -type Config,BlockDevice

// The ebsvolume package contains a packer.Builder implementation that
// builds EBS volumes for Outscale using an ephemeral instance,
package bsuvolume

import (
	"context"
	"fmt"
	"log"

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

const BuilderId = "oapi.outscale.bsuvolume"

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	osccommon.AccessConfig `mapstructure:",squash"`
	osccommon.RunConfig    `mapstructure:",squash"`

	VolumeMappings []BlockDevice `mapstructure:"bsu_volumes"`

	launchBlockDevices osccommon.BlockDevices
	ctx                interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

type EngineVarsTemplate struct {
	BuildRegion string
	SourceOMI   string
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	b.config.ctx.Funcs = osccommon.TemplateFuncs
	// Create passthrough for {{ .BuildRegion }} and {{ .SourceOMI }} variables
	// so we can fill them in later
	b.config.ctx.Data = &EngineVarsTemplate{
		BuildRegion: `{{ .BuildRegion }}`,
		SourceOMI:   `{{ .SourceOMI }} `,
	}
	err := config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Accumulate any errors
	var errs *packersdk.MultiError
	errs = packersdk.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.launchBlockDevices.Prepare(&b.config.ctx)...)

	for _, d := range b.config.VolumeMappings {
		if err := d.Prepare(&b.config.ctx); err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("OMIMapping: %s", err.Error()))
		}
	}

	b.config.launchBlockDevices, err = commonBlockDevices(b.config.VolumeMappings, &b.config.ctx)
	if err != nil {
		errs = packersdk.MultiErrorAppend(errs, err)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packersdk.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (packersdk.Artifact, error) {
	// clientConfig, err := b.config.Config()
	// if err != nil {
	// 	return nil, err
	// }

	// skipClient := &http.Client{
	// 	Transport: &http.Transport{
	// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// 	},
	// }

	oscConn := b.config.NewOSCClient()

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("osc", oscConn)
	state.Put("hook", hook)
	state.Put("ui", ui)

	log.Printf("[DEBUG] launch block devices %#v", b.config.launchBlockDevices)

	instanceStep := &osccommon.StepRunSourceVm{
		BlockDevices:                b.config.launchBlockDevices,
		Comm:                        &b.config.RunConfig.Comm,
		Ctx:                         b.config.ctx,
		Debug:                       b.config.PackerDebug,
		BsuOptimized:                b.config.BsuOptimized,
		EnableT2Unlimited:           b.config.EnableT2Unlimited,
		ExpectedRootDevice:          "bsu",
		IamVmProfile:                b.config.IamVmProfile,
		VmInitiatedShutdownBehavior: b.config.VmInitiatedShutdownBehavior,
		VmType:                      b.config.VmType,
		SourceOMI:                   b.config.SourceOmi,
		Tags:                        b.config.RunTags,
		UserData:                    b.config.UserData,
		UserDataFile:                b.config.UserDataFile,
	}

	// Build the steps
	steps := []multistep.Step{
		&osccommon.StepSourceOMIInfo{
			SourceOmi:  b.config.SourceOmi,
			OmiFilters: b.config.SourceOmiFilter,
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
			DebugKeyPath: fmt.Sprintf("oapi_%s.pem", b.config.PackerBuildName),
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
		instanceStep,
		&stepTagBSUVolumes{
			VolumeMapping: b.config.VolumeMappings,
			Ctx:           b.config.ctx,
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
			Skip:          b.config.IsSpotVm(),
			DisableStopVm: b.config.DisableStopVm,
		},
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
		Volumes:        state.Get("bsuvolumes").(BsuVolumes),
		BuilderIdValue: BuilderId,
		Conn:           oscConn,
		StateData:      map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	ui.Say(fmt.Sprintf("Created Volumes: %s", artifact))
	return artifact, nil
}
