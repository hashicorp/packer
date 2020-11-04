package docker

import (
	"context"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/packerbuilderdata"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	BuilderId       = "packer.docker"
	BuilderIdImport = "packer.post-processor.docker-import"
)

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return []string{
		"ImageSha256",
	}, warnings, nil
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

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	generatedData := &packerbuilderdata.GeneratedData{State: state}

	// Setup the driver that will talk to Docker
	state.Put("driver", driver)

	steps := []multistep.Step{
		&StepTempDir{},
		&StepPull{},
		&StepRun{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost(b.config.Comm.Host()),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
			CustomConnect: map[string]multistep.Step{
				"docker":                 &StepConnectDocker{},
				"dockerWindowsContainer": &StepConnectDocker{},
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
		steps = append(steps,
			new(StepCommit),
			&StepSetGeneratedData{ // Adds ImageSha256 variable available after StepCommit
				GeneratedData: generatedData,
			})
	} else if b.config.ExportPath != "" {
		log.Printf("[DEBUG] Container will be exported to %s", b.config.ExportPath)
		steps = append(steps, new(StepExport))
	} else {
		return nil, errArtifactNotUsed
	}

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
			StateData:      map[string]interface{}{"generated_data": state.Get("generated_data")},
		}
	} else {
		artifact = &ExportArtifact{
			path:      b.config.ExportPath,
			StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
		}
	}

	return artifact, nil
}
