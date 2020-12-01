//go:generate mapstructure-to-hcl2 -type Config,ImageFilter,ImageFilterOptions

// The openstack package contains a packer.Builder implementation that
// builds Images for openstack.

package openstack

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
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

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
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
	errs = packersdk.MultiErrorAppend(errs, b.config.ImageConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	if b.config.ImageConfig.ImageDiskFormat != "" && !b.config.RunConfig.UseBlockStorageVolume {
		return nil, nil, fmt.Errorf("use_blockstorage_volume must be true if image_disk_format is specified.")
	}

	// By default, instance name is same as image name
	if b.config.InstanceName == "" {
		b.config.InstanceName = b.config.ImageName
	}

	packersdk.LogSecretFilter.Set(b.config.Password)
	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	if b.config.PackerDebug {
		b.config.enableDebug(ui)
	}

	computeClient, err := b.config.computeV2Client()
	if err != nil {
		return nil, fmt.Errorf("Error initializing compute client: %s", err)
	}

	imageClient, err := b.config.imageV2Client()
	if err != nil {
		return nil, fmt.Errorf("Error initializing image client: %s", err)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
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
			SourceImage:               b.config.RunConfig.SourceImage,
			SourceImageName:           b.config.RunConfig.SourceImageName,
			ExternalSourceImageURL:    b.config.RunConfig.ExternalSourceImageURL,
			ExternalSourceImageFormat: b.config.RunConfig.ExternalSourceImageFormat,
			SourceImageOpts:           b.config.RunConfig.sourceImageOpts,
			SourceMostRecent:          b.config.SourceImageFilters.MostRecent,
			SourceProperties:          b.config.SourceImageFilters.Filters.Properties,
		},
		&StepDiscoverNetwork{
			Networks:              b.config.Networks,
			NetworkDiscoveryCIDRs: b.config.NetworkDiscoveryCIDRs,
			Ports:                 b.config.Ports,
		},
		&StepCreateVolume{
			UseBlockStorageVolume:  b.config.UseBlockStorageVolume,
			VolumeName:             b.config.VolumeName,
			VolumeType:             b.config.VolumeType,
			VolumeAvailabilityZone: b.config.VolumeAvailabilityZone,
		},
		&StepRunSourceServer{
			Name:                  b.config.InstanceName,
			SecurityGroups:        b.config.SecurityGroups,
			AvailabilityZone:      b.config.AvailabilityZone,
			UserData:              b.config.UserData,
			UserDataFile:          b.config.UserDataFile,
			ConfigDrive:           b.config.ConfigDrive,
			InstanceMetadata:      b.config.InstanceMetadata,
			UseBlockStorageVolume: b.config.UseBlockStorageVolume,
			ForceDelete:           b.config.ForceDelete,
		},
		&StepGetPassword{
			Debug: b.config.PackerDebug,
			Comm:  &b.config.RunConfig.Comm,
		},
		&StepWaitForRackConnect{
			Wait: b.config.RackconnectWait,
		},
		&StepAllocateIp{
			FloatingIPNetwork:     b.config.FloatingIPNetwork,
			FloatingIP:            b.config.FloatingIP,
			ReuseIPs:              b.config.ReuseIPs,
			InstanceFloatingIPNet: b.config.InstanceFloatingIPNet,
		},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: CommHost(
				b.config.RunConfig.Comm.SSHHost,
				computeClient,
				b.config.SSHInterface,
				b.config.SSHIPVersion),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.RunConfig.Comm,
		},
		&StepStopServer{},
		&StepDetachVolume{
			UseBlockStorageVolume: b.config.UseBlockStorageVolume,
		},
		&stepCreateImage{
			UseBlockStorageVolume: b.config.UseBlockStorageVolume,
		},
		&stepUpdateImageTags{},
		&stepUpdateImageVisibility{},
		&stepAddImageMembers{},
		&stepUpdateImageMinDisk{},
	}

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

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
		Client:         imageClient,
		StateData:      map[string]interface{}{"generated_data": state.Get("generated_data")},
	}

	return artifact, nil
}
