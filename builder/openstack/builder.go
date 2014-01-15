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

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	AccessConfig        `mapstructure:",squash"`
	ImageConfig         `mapstructure:",squash"`
	RunConfig           `mapstructure:",squash"`

	Tpl *packer.ConfigTemplate
}

type Builder struct {
	Config Config
	Runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	md, err := common.DecodeConfig(&b.Config, raws...)
	if err != nil {
		return nil, err
	}

	b.Config.Tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.Config.Tpl.UserVars = b.Config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.Config.AccessConfig.Prepare(b.Config.Tpl)...)
	errs = packer.MultiErrorAppend(errs, b.Config.ImageConfig.Prepare(b.Config.Tpl)...)
	errs = packer.MultiErrorAppend(errs, b.Config.RunConfig.Prepare(b.Config.Tpl)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	log.Println(common.ScrubConfig(b.Config, b.Config.Password))
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	auth, err := b.Config.AccessConfig.Auth()
	if err != nil {
		return nil, err
	}
	api := &gophercloud.ApiCriteria{
		Region:    b.Config.AccessConfig.Region(),
		UrlChoice: gophercloud.PublicURL,
	}
	
	csp, err := gophercloud.ServersApi(auth, *api)
	if err != nil {
		log.Printf("Region: %s", b.Config.AccessConfig.Region())
		return nil, err
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.Config)
	state.Put("csp", csp)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&StepKeyPair{
			Debug:        b.Config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("os_%s.pem", b.Config.PackerBuildName),
		},
		&StepRunSourceServer{
			Name:        b.Config.ImageName,
			Flavor:      b.Config.Flavor,
			SourceImage: b.Config.SourceImage,
		},
		&common.StepConnectSSH{
			SSHAddress:     SSHAddress(csp, b.Config.SSHPort),
			SSHConfig:      SSHConfig(b.Config.SSHUsername),
			SSHWaitTimeout: b.Config.SSHTimeout(),
		},
		&common.StepProvision{},
		&StepCreateImage{},
	}

	// Run!
	if b.Config.PackerDebug {
		b.Runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.Runner = &multistep.BasicRunner{Steps: steps}
	}

	b.Runner.Run(state)

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
	if b.Runner != nil {
		log.Println("Cancelling the step Runner...")
		b.Runner.Cancel()
	}
}
