
package rackspace

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"github.com/mitchellh/packer/builder/openstack"
	"log"
)

// The unique ID for this builder
const BuilderId = "mitchellh.rackspace"

type Builder struct {
	builder openstack.Builder
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	// Prepare wraps the openstack prepare function
	return b.builder.Prepare(raws...)
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	auth, err := b.builder.Config.AccessConfig.Auth()
	if err != nil {
		return nil, err
	}
	api := &gophercloud.ApiCriteria{ 
		Name:      "cloudServersOpenStack",
		Region:    b.builder.Config.AccessConfig.Region(),
		VersionId: "2",
		UrlChoice: gophercloud.PublicURL,
	}
	csp, err := gophercloud.ServersApi(auth, *api)
	if err != nil {
		log.Printf("Region: %s", b.builder.Config.AccessConfig.Region())
		return nil, err
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.builder.Config)
	state.Put("csp", csp)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&openstack.StepKeyPair{
			Debug:        b.builder.Config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("os_%s.pem", b.builder.Config.PackerBuildName),
		},
		&openstack.StepRunSourceServer{
			Name:        b.builder.Config.ImageName,
			Flavor:      b.builder.Config.Flavor,
			SourceImage: b.builder.Config.SourceImage,
		},
		&common.StepConnectSSH{
			SSHAddress:     openstack.SSHAddress(csp, b.builder.Config.SSHPort),
			SSHConfig:      openstack.SSHConfig(b.builder.Config.SSHUsername),
			SSHWaitTimeout: b.builder.Config.SSHTimeout(),
		},
		&common.StepProvision{},
		&openstack.StepCreateImage{},
	}

	// Run!
	if b.builder.Config.PackerDebug {
		b.builder.Runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.builder.Runner = &multistep.BasicRunner{Steps: steps}
	}

	b.builder.Runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no images, then just return
	if _, ok := state.GetOk("image"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &openstack.Artifact{
		ImageId:        state.Get("image").(string),
		BuilderIdValue: BuilderId,
		Conn:           csp,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	// wraps the openstack cancel function 
	b.builder.Cancel()
}
