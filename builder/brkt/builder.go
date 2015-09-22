package brkt

import (
	"log"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "brkt"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	AccessConfig        `mapstructure:",squash"`
	WorkloadConfig      `mapstructure:",squash"`
	ImageConfig         `mapstructure:",squash"`
	MachineTypeConfig   `mapstructure:",squash"`

	// SSH settings
	Comm communicator.Config `mapstructure:",squash"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	api    *brkt.API
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

	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ImageConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.WorkloadConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.Comm.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.MachineTypeConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	log.Println(common.ScrubConfig(b.config, b.config.AccessToken, b.config.MacKey))

	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	api := brkt.NewAPIForCustomPortal(b.config.AccessToken, b.config.MacKey, b.config.PortalUrl)

	state := new(multistep.BasicStateBag)
	state.Put("api", api)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&stepGetMachineType{
			MinCpuCores:   b.config.MachineTypeConfig.MinCpuCores,
			MinRam:        b.config.MachineTypeConfig.MinRam,
			AvatarEnabled: b.config.WorkloadConfig.MetavisorEnabled,
			MachineType:   b.config.MachineTypeConfig.MachineType,
		},
		&stepDeployInstance{
			ImageDefinition:  b.config.WorkloadConfig.ImageDefinition,
			BillingGroup:     b.config.WorkloadConfig.BillingGroup,
			Zone:             b.config.WorkloadConfig.Zone,
			CloudConfig:      b.config.WorkloadConfig.CloudConfig,
			SecurityGroup:    b.config.WorkloadConfig.SecurityGroup,
			MetavisorEnabled: b.config.WorkloadConfig.MetavisorEnabled,
		},
		&stepLoadKeyFile{
			PrivateKeyFile: b.config.Comm.SSHPrivateKey,
		},
		&communicator.StepConnect{
			Config: &b.config.Comm,
			Host:   SSHost,

			SSHConfig: SSHConfig(b.config.Comm.SSHUsername),
		},
		&common.StepProvision{},
		&stepCreateImage{
			ImageName: b.config.ImageConfig.ImageName,
		},
	}

	b.runner = &multistep.BasicRunner{Steps: steps}

	b.runner.Run(state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	imageId := state.Get("imageId").(string)
	imageName := state.Get("imageName").(string)

	artifact := Artifact{
		ImageId:        imageId,
		ImageName:      imageName,
		BuilderIdValue: BuilderId,
		ApiClient:      api.ApiClient,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
