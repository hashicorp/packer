// The digitalocean package contains a packer.Builder implementation
// that builds DigitalOcean images (snapshots).

package digitalocean

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

// The unique id for the builder
const BuilderId = "pearkes.digitalocean"

// Configuration tells the builder the credentials
// to use while communicating with DO and describes the image
// you are creating
type config struct {
	// Credentials
	ClientID string `mapstructure:"client_id"`
	APIKey   string `mapstructure:"api_key"`

	RegionID    uint   `mapstructure:"region_id"`
	SizeID      uint   `mapstructure:"size_id"`
	ImageID     uint   `mapstructure:"image_id"`
	SSHUsername string `mapstructure:"ssh_username"`
	SSHPort     uint   `mapstructure:"ssh_port"`

	// Configuration for the image being built
	SnapshotName string `mapstructure:"snapshot_name"`

	RawSSHTimeout string `mapstructure:"ssh_timeout"`
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raw interface{}) error {
	if err := mapstructure.Decode(raw, &b.config); err != nil {
		return err
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
		// Default to base image "Ubuntu 12.04 x64 Server"
		b.config.ImageID = 2676
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
		// Default to packer-{{ unix timestamp (utc) }}
		b.config.SnapshotName = "packer-{{.CreateTime}}"
	}

	if b.config.RawSSHTimeout == "" {
		// Default to 1 minute timeouts
		b.config.RawSSHTimeout = "1m"
	}

	// A list of errors on the configuration
	errs := make([]error, 0)

	// Required configurations that will display errors if not set
	//
	if b.config.ClientId == "" {
		errs = append(errs, errors.New("a client_id must be specified"))
	}

	if b.config.APIKey == "" {
		errs = append(errs, errors.New("an api_key must be specified"))
	}

	b.config.SSHTimeout, err = time.ParseDuration(b.config.RawSSHTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	log.Printf("Config: %+v", b.config)
	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {

}
