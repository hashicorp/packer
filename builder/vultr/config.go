package vultr

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`
	ctx                 interpolate.Context

	APIKey string `mapstructure:"api_key"`

	Description string `mapstructure:"snapshot_description"`
	RegionID    int    `mapstructure:"region_id"`
	PlanID      int    `mapstructure:"plan_id"`
	OSID        int    `mapstructure:"os_id"`
	SnapshotID  string `mapstructure:"snapshot_id"`
	ISOID       int    `mapstructure:"iso_id"`
	AppID       string `mapstructure:"app_id"`

	EnableIPV6           bool     `mapstructure:"enable_ipv6"`
	EnablePrivateNetwork bool     `mapstructure:"enable_private_network"`
	ScriptID             string   `mapstructure:"script_id"`
	SSHKeyIDs            []string `mapstructure:"ssh_key_ids"`
	Label                string   `mapstructure:"instance_label"`
	UserData             string   `mapstructure:"userdata"`
	Hostname             string   `mapstructure:"hostname"`
	Tag                  string   `mapstructure:"tag"`

	RawStateTimeout string `mapstructure:"state_timeout"`

	stateTimeout time.Duration
	interCtx     interpolate.Context
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

	if c.APIKey == "" {
		// Default to environment variable for api_key, if it exists
		c.APIKey = os.Getenv("VULTR_API_KEY")
		if c.APIKey == "" {
			errs = packer.MultiErrorAppend(errs, errors.New("vultr api_key is required"))
		}
	}

	if c.Description == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("unable to render snapshot description: %s", err))
		} else {
			c.Description = def
		}
	}

	if c.Label == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("unable to render label: %s", err))
		} else {
			c.Label = def
		}
	}

	if c.RegionID == 0 {
		errs = packer.MultiErrorAppend(errs, errors.New("region_id is required"))
	}

	if c.PlanID == 0 {
		errs = packer.MultiErrorAppend(errs, errors.New("plan_id is required"))
	}

	if (c.AppID != "" && c.SnapshotID != "") || (c.AppID != "" && c.ISOID != 0) || (c.SnapshotID != "" && c.ISOID != 0) {
		errs = packer.MultiErrorAppend(errs, errors.New("you can only set one of the following: `app_id`, `snapshot_id`, `iso_id`"))
	}

	if c.AppID != "" {
		c.OSID = AppOSID
	}
	if c.SnapshotID != "" {
		c.OSID = SnapshotOSID
	}
	if c.ISOID != 0 {
		c.OSID = CustomOSID
	}

	if c.OSID == 0 {
		errs = packer.MultiErrorAppend(errs, errors.New("os_id is required"))
	}

	if (c.OSID == SnapshotOSID || c.OSID == CustomOSID) && c.Comm.SSHPassword == "" && c.Comm.SSHPrivateKeyFile == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("either `ssh_password` or `ssh_private_key_file` must be defined for snapshot or custom OS"))
	}

	if c.RawStateTimeout == "" {
		c.stateTimeout = 5 * time.Minute
	} else {
		if stateTimeout, err := time.ParseDuration(c.RawStateTimeout); err == nil {
			c.stateTimeout = stateTimeout
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("unable to parse state timeout: %s", err))
		}
	}

	if es := c.Comm.Prepare(&c.interCtx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(c.APIKey)

	return c, nil, nil
}
