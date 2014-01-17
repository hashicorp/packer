// The openstack package contains a packer.Builder implementation that
// builds Images for openstack.

package openstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"log"
)

// The unique ID for this builder
const BuilderId = "mitchellh.openstack"

type config struct {
	common.PackerConfig `mapstructure:",squash"`
	AccessConfig        `mapstructure:",squash"`
	ImageConfig         `mapstructure:",squash"`
	RunConfig           `mapstructure:",squash"`

	tpl *packer.ConfigTemplate
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.ImageConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(b.config.tpl)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	log.Println(common.ScrubConfig(b.config, b.config.Password))
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	auth, err := b.config.AccessConfig.Auth()
	if err != nil {
		return nil, err
	}
	//fetches the api requisites from gophercloud for the appropriate
	//openstack variant
	api, err := gophercloud.PopulateApi(b.config.RunConfig.OpenstackProvider)
	if err != nil {
		return nil, err
	}
	api.Region = b.config.AccessConfig.Region()

	csp, err := gophercloud.ServersApi(auth, api)
	if err != nil {
		log.Printf("Region: %s", b.config.AccessConfig.Region())
		return nil, err
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("csp", csp)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&StepKeyPair{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("os_%s.pem", b.config.PackerBuildName),
		},
		&StepRunSourceServer{
			Name:           b.config.ImageName,
			Flavor:         b.config.Flavor,
			SourceImage:    b.config.SourceImage,
			SecurityGroups: b.config.SecurityGroups,
		},
		&StepAllocateIp{
			FloatingIpPool: b.config.FloatingIpPool,
			FloatingIp:     b.config.FloatingIp,
		},
		&common.StepConnectSSH{
			SSHAddress:     SSHAddress(csp, b.config.SSHPort),
			SSHConfig:      SSHConfig(b.config.SSHUsername),
			SSHWaitTimeout: b.config.SSHTimeout(),
		},
		&common.StepProvision{},
		&stepCreateImage{},
	}

	// Run!
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

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
		Conn:           csp,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
