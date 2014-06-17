// The digitalocean package contains a packer.Builder implementation
// that builds DigitalOcean images (snapshots).

package digitalocean

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"time"
)

// see https://api.digitalocean.com/images/?client_id=[client_id]&api_key=[api_key]
// name="Ubuntu 12.04.4 x64", id=3101045,
const DefaultImage = "ubuntu-12-04-x64"

// see https://api.digitalocean.com/regions/?client_id=[client_id]&api_key=[api_key]
// name="New York", id=1
const DefaultRegion = "nyc1"

// see https://api.digitalocean.com/sizes/?client_id=[client_id]&api_key=[api_key]
// name="512MB", id=66 (the smallest droplet size)
const DefaultSize = "512mb"

// The unique id for the builder
const BuilderId = "pearkes.digitalocean"

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

	Region string `mapstructure:"region"`
	Size   string `mapstructure:"size"`
	Image  string `mapstructure:"image"`

	PrivateNetworking bool   `mapstructure:"private_networking"`
	SnapshotName      string `mapstructure:"snapshot_name"`
	DropletName       string `mapstructure:"droplet_name"`
	SSHUsername       string `mapstructure:"ssh_username"`
	SSHPort           uint   `mapstructure:"ssh_port"`

	RawSSHTimeout   string `mapstructure:"ssh_timeout"`
	RawStateTimeout string `mapstructure:"state_timeout"`

	// These are unexported since they're set by other fields
	// being set.
	sshTimeout   time.Duration
	stateTimeout time.Duration

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

	// Optional configuration with defaults
	if b.config.APIKey == "" {
		// Default to environment variable for api_key, if it exists
		b.config.APIKey = os.Getenv("DIGITALOCEAN_API_KEY")
	}

	if b.config.ClientID == "" {
		// Default to environment variable for client_id, if it exists
		b.config.ClientID = os.Getenv("DIGITALOCEAN_CLIENT_ID")
	}

	if b.config.Region == "" {
		if b.config.RegionID != 0 {
			b.config.Region = fmt.Sprintf("%v", b.config.RegionID)
		} else {
			b.config.Region = DefaultRegion
		}
	}

	if b.config.Size == "" {
		if b.config.SizeID != 0 {
			b.config.Size = fmt.Sprintf("%v", b.config.SizeID)
		} else {
			b.config.Size = DefaultSize
		}
	}

	if b.config.Image == "" {
		if b.config.ImageID != 0 {
			b.config.Image = fmt.Sprintf("%v", b.config.ImageID)
		} else {
			b.config.Image = DefaultImage
		}
	}

	if b.config.SnapshotName == "" {
		// Default to packer-{{ unix timestamp (utc) }}
		b.config.SnapshotName = "packer-{{timestamp}}"
	}

	if b.config.DropletName == "" {
		// Default to packer-[time-ordered-uuid]
		b.config.DropletName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
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

	if b.config.RawSSHTimeout == "" {
		// Default to 1 minute timeouts
		b.config.RawSSHTimeout = "1m"
	}

	if b.config.RawStateTimeout == "" {
		// Default to 6 minute timeouts waiting for
		// desired state. i.e waiting for droplet to become active
		b.config.RawStateTimeout = "6m"
	}

	templates := map[string]*string{
		"client_id":     &b.config.ClientID,
		"api_key":       &b.config.APIKey,
		"snapshot_name": &b.config.SnapshotName,
		"droplet_name":  &b.config.DropletName,
		"ssh_username":  &b.config.SSHUsername,
		"ssh_timeout":   &b.config.RawSSHTimeout,
		"state_timeout": &b.config.RawStateTimeout,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
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

	stateTimeout, err := time.ParseDuration(b.config.RawStateTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing state_timeout: %s", err))
	}
	b.config.stateTimeout = stateTimeout

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	common.ScrubConfig(b.config, b.config.ClientID, b.config.APIKey)
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Initialize the DO API client
	client := DigitalOceanClient{}.New(b.config.ClientID, b.config.APIKey)

	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		new(stepCreateSSHKey),
		new(stepCreateDroplet),
		new(stepDropletInfo),
		&common.StepConnectSSH{
			SSHAddress:     sshAddress,
			SSHConfig:      sshConfig,
			SSHWaitTimeout: 5 * time.Minute,
		},
		new(common.StepProvision),
		new(stepShutdown),
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
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("snapshot_name"); !ok {
		log.Println("Failed to find snapshot_name in state. Bug?")
		return nil, nil
	}

	sregion := state.Get("region")

	var region string

	if sregion != nil {
		region = sregion.(string)
	} else {
		region = fmt.Sprintf("%v", state.Get("region_id").(uint))
	}

	found_region, err := client.Region(region)

	if err != nil {
		return nil, err
	}

	artifact := &Artifact{
		snapshotName: state.Get("snapshot_name").(string),
		snapshotId:   state.Get("snapshot_image_id").(uint),
		regionName:   found_region.Name,
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
