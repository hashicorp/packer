//go:generate mapstructure-to-hcl2 -type Config,BlockDevice

// The ebsvolume package contains a packer.Builder implementation that
// builds EBS volumes for Outscale using an ephemeral instance,
package bsuvolume

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/hcl/v2/hcldec"
	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
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

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config.ctx.Funcs = osccommon.TemplateFuncs
	// Create passthrough for {{ .BuildRegion }} and {{ .SourceOMI }} variables
	// so we can fill them in later
	b.config.ctx.Data = &EngineVarsTemplate{
		BuildRegion: `{{ .BuildRegion }}`,
		SourceOMI:   `{{ .SourceOMI }} `,
	}
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.launchBlockDevices.Prepare(&b.config.ctx)...)

	for _, d := range b.config.VolumeMappings {
		if err := d.Prepare(&b.config.ctx); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("OMIMapping: %s", err.Error()))
		}
	}

	b.config.launchBlockDevices, err = commonBlockDevices(b.config.VolumeMappings, &b.config.ctx)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
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

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("oapi", oapiconn)
	state.Put("hook", hook)
	state.Put("ui", ui)

	log.Printf("[DEBUG] launch block devices %#v", b.config.launchBlockDevices)

	instanceStep := &osccommon.StepRunSourceVm{
		AssociatePublicIpAddress:    b.config.AssociatePublicIpAddress,
		BlockDevices:                b.config.launchBlockDevices,
		Comm:                        &b.config.RunConfig.Comm,
		Ctx:                         b.config.ctx,
		Debug:                       b.config.PackerDebug,
		BsuOptimized:                b.config.BsuOptimized,
		EnableT2Unlimited:           b.config.EnableT2Unlimited,
		ExpectedRootDevice:          "ebs",
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
			Host: osccommon.SSHHost(
				oapiconn,
				b.config.SSHInterface),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.RunConfig.Comm,
		},
		&osccommon.StepStopBSUBackedVm{
			Skip:          b.config.IsSpotVm(),
			DisableStopVm: b.config.DisableStopVm,
		},
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// Build the artifact and return it
	artifact := &Artifact{
		Volumes:        state.Get("bsuvolumes").(BsuVolumes),
		BuilderIdValue: BuilderId,
		Conn:           oapiconn,
	}
	ui.Say(fmt.Sprintf("Created Volumes: %s", artifact))
	return artifact, nil
}
