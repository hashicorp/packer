// The openstack package contains a packer.Builder implementation that
// builds Images for openstack.

package openstack

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// The unique ID for this builder
const BuilderId = "mitchellh.openstack"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	AccessConfig `mapstructure:",squash"`
	ImageConfig  `mapstructure:",squash"`
	RunConfig    `mapstructure:",squash"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
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
	errs = packer.MultiErrorAppend(errs, b.config.ImageConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	// By default, instance name is same as image name
	if b.config.InstanceName == "" {
		b.config.InstanceName = b.config.ImageName
	}

	packer.LogSecretFilter.Set(b.config.Password)
	log.Println(b.config)
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	computeClient, err := b.config.computeV2Client()
	if err != nil {
		return nil, fmt.Errorf("Error initializing compute client: %s", err)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&StepLoadFlavor{
			Flavor: b.config.Flavor,
		},
		&StepKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.Comm,
			DebugKeyPath: fmt.Sprintf("os_%s.pem", b.config.PackerBuildName),
		},
		&StepSourceImageInfo{
			SourceImage:      b.config.RunConfig.SourceImage,
			SourceImageName:  b.config.RunConfig.SourceImageName,
			SourceImageOpts:  b.config.RunConfig.sourceImageOpts,
			SourceMostRecent: b.config.SourceImageFilters.MostRecent,
		},
		&StepCreateVolume{
			UseBlockStorageVolume:  b.config.UseBlockStorageVolume,
			SourceImage:            b.config.SourceImage,
			VolumeName:             b.config.VolumeName,
			VolumeType:             b.config.VolumeType,
			VolumeAvailabilityZone: b.config.VolumeAvailabilityZone,
		},
		&StepRunSourceServer{
			Name:                  b.config.InstanceName,
			SourceImage:           b.config.SourceImage,
			SourceImageName:       b.config.SourceImageName,
			SecurityGroups:        b.config.SecurityGroups,
			Networks:              b.config.Networks,
			Ports:                 b.config.Ports,
			AvailabilityZone:      b.config.AvailabilityZone,
			UserData:              b.config.UserData,
			UserDataFile:          b.config.UserDataFile,
			ConfigDrive:           b.config.ConfigDrive,
			InstanceMetadata:      b.config.InstanceMetadata,
			UseBlockStorageVolume: b.config.UseBlockStorageVolume,
			Comm:                  &b.config.Comm,
		},
		&StepGetPassword{
			Debug: b.config.PackerDebug,
			Comm:  &b.config.RunConfig.Comm,
		},
		&StepWaitForRackConnect{
			Wait: b.config.RackconnectWait,
		},
		&StepAllocateIp{
			FloatingIPNetwork: b.config.FloatingIPNetwork,
			FloatingIP:        b.config.FloatingIP,
			ReuseIPs:          b.config.ReuseIPs,
		},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: CommHost(
				computeClient,
				b.config.Comm.SSHInterface,
				b.config.Comm.SSHIPVersion),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.RunConfig.Comm,
		},
		&StepStopServer{},
		&StepDetachVolume{
			UseBlockStorageVolume: b.config.UseBlockStorageVolume,
		},
		&stepCreateImage{
			UseBlockStorageVolume: b.config.UseBlockStorageVolume,
		},
		&stepUpdateImageVisibility{},
		&stepAddImageMembers{},
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no images, then just return
	if _, ok := state.GetOk("image"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &Artifact{
		ImageId:        state.Get("image").(string),
		BuilderIdValue: BuilderId,
		Client:         computeClient,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
