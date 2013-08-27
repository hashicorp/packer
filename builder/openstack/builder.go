// The openstack package contains a packer.Builder implementation that
// builds Images for openstack.

package openstack

import (
	//"fmt"
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

func (b *Builder) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.ImageConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(b.config.tpl)...)

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	log.Printf("Config: %+v", b.config)
	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	auth, err := b.config.AccessConfig.Auth()
	if err != nil {
		return nil, err
	}

	// Setup the state bag and initial state for the steps
	state := make(map[string]interface{})
	state["config"] = b.config
	state["accessor"] = auth
	api := &gophercloud.ApiCriteria{
		Name:      "cloudServersOpenStack",
		Region:    "DFW",
		VersionId: "2",
		UrlChoice: gophercloud.PublicURL,
	}
	state["api"] = api

	state["hook"] = hook
	state["ui"] = ui

	// Build the steps
	steps := []multistep.Step{
		&StepKeyPair{},
		&StepRunSourceServer{
			Name:        b.config.ImageName,
			Flavor:      b.config.Flavor,
			SourceImage: b.config.SourceImage,
		},
		&common.StepConnectSSH{
			SSHAddress:     SSHAddress(&auth, api, b.config.SSHPort),
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
	if rawErr, ok := state["error"]; ok {
		return nil, rawErr.(error)
	}

	// XXX - add artifact
	return nil, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
