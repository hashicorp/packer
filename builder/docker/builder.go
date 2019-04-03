package docker

import (
	"context"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	BuilderId       = "packer.docker"
	BuilderIdImport = "packer.post-processor.docker-import"
)

type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	return warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	driver := &DockerDriver{Ctx: &b.config.ctx, Ui: ui}
	if err := driver.Verify(); err != nil {
		return nil, err
	}

	version, err := driver.Version()
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Docker version: %s", version.String())

	steps := []multistep.Step{
		&StepTempDir{},
		&StepPull{},
		&StepRun{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
			CustomConnect: map[string]multistep.Step{
				"docker": &StepConnectDocker{},
			},
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
	}

	if b.config.Discard {
		log.Print("[DEBUG] Container will be discarded")
	} else if b.config.Commit {
		log.Print("[DEBUG] Container will be committed")
		steps = append(steps, new(StepCommit))
	} else if b.config.ExportPath != "" {
		log.Printf("[DEBUG] Container will be exported to %s", b.config.ExportPath)
		steps = append(steps, new(StepExport))
	} else {
		return nil, errArtifactNotUsed
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Setup the driver that will talk to Docker
	state.Put("driver", driver)

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If it was cancelled, then just return
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, nil
	}

	// No errors, must've worked
	var artifact packer.Artifact
	if b.config.Commit {
		artifact = &ImportArtifact{
			IdValue:        state.Get("image_id").(string),
			BuilderIdValue: BuilderIdImport,
			Driver:         driver,
		}
	} else {
		artifact = &ExportArtifact{path: b.config.ExportPath}
	}

	return artifact, nil
}
