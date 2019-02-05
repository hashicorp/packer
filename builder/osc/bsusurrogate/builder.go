// The bsusurrogate package contains a packer.Builder implementation that
// builds a new EBS-backed OMI using an ephemeral instance.
package bsusurrogate

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
)

const BuilderId = "digitalonus.osc.bsusurrogate"

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

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {

	b.config.ctx.Funcs = osccommon.TemplateFuncs

	err := config.Decode(&b.config, &config.DecodeOpts{
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
		return nil, err
	}

	if b.config.PackerConfig.PackerForce {
		b.config.OMIForceDeregister = true
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs,
		b.config.OMIConfig.Prepare(&b.config.AccessConfig, &b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.BlockDevices.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RootDevice.Prepare(&b.config.ctx)...)

	if b.config.OMIVirtType == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("omi_virtualization_type is required."))
	}

	foundRootVolume := false
	for _, launchDevice := range b.config.BlockDevices.LaunchMappings {
		if launchDevice.DeviceName == b.config.RootDevice.SourceDeviceName {
			foundRootVolume = true
		}
	}

	if !foundRootVolume {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("no volume with name '%s' is found", b.config.RootDevice.SourceDeviceName))
	}

	//TODO: Chek if this is necessary
	if b.config.IsSpotVm() && ((b.config.OMIENASupport != nil && *b.config.OMIENASupport) || b.config.OMISriovNetSupport) {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Spot instances do not support modification, which is required "+
				"when either `ena_support` or `sriov_support` are set. Please ensure "+
				"you use an OMI that already has either SR-IOV or ENA enabled."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return nil, nil

}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
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
	state.Put("clientConfig", clientConfig)
	state.Put("hook", hook)
	state.Put("ui", ui)

	//VMStep

	omiDevices := b.config.BuildOMIDevices()
	launchDevices := b.config.BuildLaunchDevices()

	steps := []multistep.Step{
		&osccommon.StepPreValidate{
			DestOmiName:     b.config.OMIName,
			ForceDeregister: b.config.OMIForceDeregister,
		},
		&osccommon.StepSourceOMIInfo{
			SourceOmi:                b.config.SourceOmi,
			EnableOMISriovNetSupport: b.config.OMISriovNetSupport,
			EnableOMIENASupport:      b.config.OMIENASupport,
			OmiFilters:               b.config.SourceOmiFilter,
			OMIVirtType:              b.config.OMIVirtType, //TODO: Remove if it is not used
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
		&osccommon.StepCleanupVolumes{
			BlockDevices: b.config.BlockDevices,
		},
		&osccommon.StepRunSourceVm{
			AssociatePublicIpAddress:    b.config.AssociatePublicIpAddress,
			BlockDevices:                b.config.BlockDevices,
			Comm:                        &b.config.RunConfig.Comm,
			Ctx:                         b.config.ctx,
			Debug:                       b.config.PackerDebug,
			BsuOptimized:                b.config.BsuOptimized,
			EnableT2Unlimited:           b.config.EnableT2Unlimited,
			ExpectedRootDevice:          "ebs", // should it be bsu
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
			Host: osccommon.SSHHost(
				oapiconn,
				b.config.Comm.SSHInterface),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.RunConfig.Comm,
		},
		&osccommon.StepStopBSUBackedVm{
			Skip:          false,
			DisableStopVm: b.config.DisableStopVm,
		},
		&osccommon.StepUpdateBSUBackedVm{
			EnableAMISriovNetSupport: b.config.OMISriovNetSupport,
			EnableAMIENASupport:      b.config.OMIENASupport,
		},
		&StepSnapshotVolumes{
			LaunchDevices: launchDevices,
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
			LaunchDevices: launchDevices,
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

	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	//Build the artifact
	// if omis, ok := state.GetOk("omis"); ok {
	// 	// Build the artifact and return it
	// 	artifact := &awscommon.Artifact{
	// 		Amis:           omis.(map[string]string),
	// 		BuilderIdValue: BuilderId,
	// 		Session:        session,
	// 	}

	// 	return artifact, nil
	// }

	return nil, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
