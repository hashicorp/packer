// The digitalocean package contains a packer.Builder implementation
// that builds DigitalOcean images (snapshots).

package digitalocean

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"strconv"
	"text/template"
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
	ClientID string `mapstructure:"client_id"`
	APIKey   string `mapstructure:"api_key"`
	RegionID uint   `mapstructure:"region_id"`
	SizeID   uint   `mapstructure:"size_id"`
	ImageID  uint   `mapstructure:"image_id"`

	SnapshotName string
	SSHUsername  string `mapstructure:"ssh_username"`
	SSHPort      uint   `mapstructure:"ssh_port"`
	SSHTimeout   time.Duration
	EventDelay   time.Duration

	PackerDebug bool `mapstructure:"packer_debug"`

	RawSnapshotName string `mapstructure:"snapshot_name"`
	RawSSHTimeout   string `mapstructure:"ssh_timeout"`
	RawEventDelay   string `mapstructure:"event_delay"`
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) error {
	for _, raw := range raws {
		err := mapstructure.Decode(raw, &b.config)
		if err != nil {
			return err
		}
	}

	// Optional configuration with defaults
	//
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

	if b.config.RawSnapshotName == "" {
		// Default to packer-{{ unix timestamp (utc) }}
		b.config.RawSnapshotName = "packer-{{.CreateTime}}"
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

	// A list of errors on the configuration
	errs := make([]error, 0)

	// Required configurations that will display errors if not set
	//
	if b.config.ClientID == "" {
		errs = append(errs, errors.New("a client_id must be specified"))
	}

	if b.config.APIKey == "" {
		errs = append(errs, errors.New("an api_key must be specified"))
	}

	timeout, err := time.ParseDuration(b.config.RawSSHTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}
	b.config.SSHTimeout = timeout

	delay, err := time.ParseDuration(b.config.RawEventDelay)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing event_delay: %s", err))
	}
	b.config.EventDelay = delay

	// Parse the name of the snapshot
	snapNameBuf := new(bytes.Buffer)
	tData := snapshotNameData{
		strconv.FormatInt(time.Now().UTC().Unix(), 10),
	}
	t, err := template.New("snapshot").Parse(b.config.RawSnapshotName)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing snapshot_name: %s", err))
	} else {
		t.Execute(snapNameBuf, tData)
		b.config.SnapshotName = snapNameBuf.String()
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
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
		new(stepCreateSSHKey),
		new(stepCreateDroplet),
		new(stepDropletInfo),
		new(stepConnectSSH),
		new(stepProvision),
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

	return &Artifact{b.config.SnapshotName}, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
