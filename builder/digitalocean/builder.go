// The digitalocean package contains a packer.Builder implementation
// that builds DigitalOcean images (snapshots).

package digitalocean

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"time"
)

// The unique id for the builder
const BuilderId = "pearkes.digitalocean"

type snapshotNameData struct {
	CreateTime string
}

// Configuration tells the builder the credentials
// to use while communicating with DO and describes the image
// you are creating
type config struct {
	common.PackerConfig `mapstructure:",squash"`

	ClientID string `mapstructure:"client_id"`
	APIKey   string `mapstructure:"api_key"`
	RegionID uint   `mapstructure:"region_id"`
	SizeID   uint   `mapstructure:"size_id"`
	ImageID  uint   `mapstructure:"image_id"`

	SnapshotName string `mapstructure:"snapshot_name"`
	SSHUsername  string `mapstructure:"ssh_username"`
	SSHPort      uint   `mapstructure:"ssh_port"`

	RawSSHTimeout   string `mapstructure:"ssh_timeout"`
	RawEventDelay   string `mapstructure:"event_delay"`
	RawStateTimeout string `mapstructure:"state_timeout"`

	// These are unexported since they're set by other fields
	// being set.
	sshTimeout   time.Duration
	eventDelay   time.Duration
	stateTimeout time.Duration

	template *common.ConfigTemplate
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

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	// Optional configuration with defaults
	if b.config.APIKey == "" {
		// Default to environment variable for api_key, if it exists
		b.config.APIKey = os.Getenv("DIGITALOCEAN_API_KEY")
	}

	if b.config.ClientID == "" {
		// Default to environment variable for client_id, if it exists
		b.config.ClientID = os.Getenv("DIGITALOCEAN_CLIENT_ID")
	}

	if b.config.RegionID == 0 {
		// Default to Region "New York"
		b.config.RegionID = 1
	}

	if b.config.SizeID == 0 {
		// Default to 512mb, the smallest droplet size
		b.config.SizeID = 66
	}

	if b.config.ImageID == 0 {
		// Default to base image "Ubuntu 12.04 x64 Server (id: 284203)"
		b.config.ImageID = 284203
	}

	if b.config.SSHUsername == "" {
		// Default to "root". You can override this if your
		// SourceImage has a different user account then the DO default
		b.config.SSHUsername = "root"
	}

	if b.config.SSHPort == 0 {
		// Default to port 22 per DO default
		b.config.SSHPort = 22
	}

	if b.config.SnapshotName == "" {
		b.config.SnapshotName = "packer-{{timestamp}}"
	}

	if b.config.RawSSHTimeout == "" {
		// Default to 1 minute timeouts
		b.config.RawSSHTimeout = "1m"
	}

	if b.config.RawEventDelay == "" {
		// Default to 5 second delays after creating events
		// to allow DO to process
		b.config.RawEventDelay = "5s"
	}

	if b.config.RawStateTimeout == "" {
		// Default to 6 minute timeouts waiting for
		// desired state. i.e waiting for droplet to become active
		b.config.RawStateTimeout = "6m"
	}

	// Required configurations that will display errors if not set
	if b.config.ClientID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a client_id must be specified"))
	}

	if b.config.APIKey == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("an api_key must be specified"))
	}

	sshTimeout, err := time.ParseDuration(b.config.RawSSHTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}
	b.config.sshTimeout = sshTimeout

	eventDelay, err := time.ParseDuration(b.config.RawEventDelay)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing event_delay: %s", err))
	}
	b.config.eventDelay = eventDelay

	stateTimeout, err := time.ParseDuration(b.config.RawStateTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing state_timeout: %s", err))
	}
	b.config.stateTimeout = stateTimeout

	// Templating
	b.config.template, err = common.NewConfigTemplate(&b.config)
	if err != nil {
		panic(err)
	}

	err = b.config.template.Check()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	log.Printf("Config: %+v", b.config)
	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Initialize the DO API client
	client := DigitalOceanClient{}.New(b.config.ClientID, b.config.APIKey)

	// Set up the state
	state := make(map[string]interface{})
	state["config"] = b.config
	state["client"] = client
	state["hook"] = hook
	state["ui"] = ui

	// Build the steps
	steps := []multistep.Step{
		&common.StepProcessConfigTemplate{
			ConfigTemplate: b.config.template,
		},
		new(stepCreateSSHKey),
		new(stepCreateDroplet),
		new(stepDropletInfo),
		&common.StepConnectSSH{
			SSHAddress:     sshAddress,
			SSHConfig:      sshConfig,
			SSHWaitTimeout: 5 * time.Minute,
		},
		new(common.StepProvision),
		new(stepPowerOff),
		new(stepSnapshot),
	}

	// Run the steps
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

	if _, ok := state["snapshot_name"]; !ok {
		log.Println("Failed to find snapshot_name in state. Bug?")
		return nil, nil
	}

	artifact := &Artifact{
		snapshotName: state["snapshot_name"].(string),
		snapshotId:   state["snapshot_image_id"].(uint),
		client:       client,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
