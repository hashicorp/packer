package linode

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	Comm                communicator.Config `mapstructure:",squash"`

	PersonalAccessToken string `mapstructure:"linode_token"`

	Region       string   `mapstructure:"region"`
	InstanceType string   `mapstructure:"instance_type"`
	Label        string   `mapstructure:"instance_label"`
	Tags         []string `mapstructure:"instance_tags"`
	Image        string   `mapstructure:"image"`
	SwapSize     int      `mapstructure:"swap_size"`
	RootPass     string   `mapstructure:"root_pass"`
	RootSSHKey   string   `mapstructure:"root_ssh_key"`
	ImageLabel   string   `mapstructure:"image_label"`
	Description  string   `mapstructure:"image_description"`

	RawStateTimeout string `mapstructure:"state_timeout"`

	stateTimeout time.Duration
	interCtx     interpolate.Context
}

func createRandomRootPassword() (string, error) {
	rawRootPass := make([]byte, 50)
	_, err := rand.Read(rawRootPass)
	if err != nil {
		return "", fmt.Errorf("Failed to generate random password")
	}
	rootPass := base64.StdEncoding.EncodeToString(rawRootPass)
	return rootPass, nil
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)

	if err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...); err != nil {
		return nil, nil, err
	}

	var errs *packer.MultiError

	// Defaults

	if c.PersonalAccessToken == "" {
		// Default to environment variable for linode_token, if it exists
		c.PersonalAccessToken = os.Getenv("LINODE_TOKEN")
	}

	if c.ImageLabel == "" {
		if def, err := interpolate.Render("packer-{{timestamp}}", nil); err == nil {
			c.ImageLabel = def
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Unable to render image name: %s", err))
		}
	}

	if c.Label == "" {
		// Default to packer-[time-ordered-uuid]
		if def, err := interpolate.Render("packer-{{timestamp}}", nil); err == nil {
			c.Label = def
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Unable to render Linode label: %s", err))
		}
	}

	if c.RootPass == "" {
		var err error
		c.RootPass, err = createRandomRootPassword()
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Unable to generate root_pass: %s", err))
		}
	}

	if c.RawStateTimeout == "" {
		c.stateTimeout = 5 * time.Minute
	} else {
		if stateTimeout, err := time.ParseDuration(c.RawStateTimeout); err == nil {
			c.stateTimeout = stateTimeout
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Unable to parse state timeout: %s", err))
		}
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	c.Comm.SSHPassword = c.RootPass

	if c.PersonalAccessToken == "" {
		// Required configurations that will display errors if not set
		errs = packer.MultiErrorAppend(
			errs, errors.New("linode_token is required"))
	}

	if c.Region == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("region is required"))
	}

	if c.InstanceType == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("instance_type is required"))
	}

	if c.Image == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("image is required"))
	}

	if c.Tags == nil {
		c.Tags = make([]string, 0)
	}
	tagRe := regexp.MustCompile("^[[:alnum:]:_-]{1,255}$")

	for _, t := range c.Tags {
		if !tagRe.MatchString(t) {
			errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("invalid tag: %s", t)))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(c.PersonalAccessToken)
	return c, nil, nil
}
