package hcloud

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// The unique id for the builder
const BuilderId = "hcloud.builder"

type Builder struct {
	config       Config
	runner       multistep.Runner
	hcloudClient *hcloud.Client
}

var pluginVersion = "1.0.0"

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	config, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = *config
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	opts := []hcloud.ClientOption{
		hcloud.WithToken(b.config.HCloudToken),
		hcloud.WithEndpoint(b.config.Endpoint),
		hcloud.WithPollInterval(b.config.PollInterval),
		hcloud.WithApplication("hcloud-packer", pluginVersion),
	}
	b.hcloudClient = hcloud.NewClient(opts...)
	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hcloudClient", b.hcloudClient)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&stepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("ssh_key_%s.pem", b.config.PackerBuildName),
		},
		&stepCreateServer{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      getServerIP,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&stepShutdownServer{},
		&stepCreateSnapshot{},
	}
	// Run the steps
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)
	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("snapshot_name"); !ok {
		return nil, nil
	}

	artifact := &Artifact{
		snapshotName: state.Get("snapshot_name").(string),
		snapshotId:   state.Get("snapshot_id").(int),
		hcloudClient: b.hcloudClient,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
