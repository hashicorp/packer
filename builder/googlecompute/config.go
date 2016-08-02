package googlecompute

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

var reImageFamily = regexp.MustCompile(`^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`)

// Config is the configuration structure for the GCE builder. It stores
// both the publicly settable state as well as the privately generated
// state of the config object.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	AccountFile string `mapstructure:"account_file"`
	ProjectId   string `mapstructure:"project_id"`

	Address              string            `mapstructure:"address"`
	DiskName             string            `mapstructure:"disk_name"`
	DiskSizeGb           int64             `mapstructure:"disk_size"`
	DiskType             string            `mapstructure:"disk_type"`
	ImageName            string            `mapstructure:"image_name"`
	ImageDescription     string            `mapstructure:"image_description"`
	ImageFamily          string            `mapstructure:"image_family"`
	InstanceName         string            `mapstructure:"instance_name"`
	MachineType          string            `mapstructure:"machine_type"`
	Metadata             map[string]string `mapstructure:"metadata"`
	Network              string            `mapstructure:"network"`
	OmitExternalIP       bool              `mapstructure:"omit_external_ip"`
	Preemptible          bool              `mapstructure:"preemptible"`
	RawStateTimeout      string            `mapstructure:"state_timeout"`
	Region               string            `mapstructure:"region"`
	SourceImage          string            `mapstructure:"source_image"`
	SourceImageProjectId string            `mapstructure:"source_image_project_id"`
	StartupScriptFile    string            `mapstructure:"startup_script_file"`
	Subnetwork           string            `mapstructure:"subnetwork"`
	Tags                 []string          `mapstructure:"tags"`
	UseInternalIP        bool              `mapstructure:"use_internal_ip"`
	Zone                 string            `mapstructure:"zone"`

	account         accountFile
	privateKeyBytes []byte
	stateTimeout    time.Duration
	ctx             interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	var errs *packer.MultiError

	// Set defaults.
	if c.Network == "" {
		c.Network = "default"
	}

	if c.DiskSizeGb == 0 {
		c.DiskSizeGb = 10
	}

	if c.DiskType == "" {
		c.DiskType = "pd-standard"
	}

	if c.ImageDescription == "" {
		c.ImageDescription = "Created by Packer"
	}

	if c.ImageName == "" {
		img, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Unable to parse image name: %s ", err))
		} else {
			c.ImageName = img
		}
	}

	if len(c.ImageFamily) > 63 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("Invalid image family: Must not be longer than 63 characters"))
	}

	if c.ImageFamily != "" {
		if !reImageFamily.MatchString(c.ImageFamily) {
			errs = packer.MultiErrorAppend(errs,
				errors.New("Invalid image family: The first character must be a lowercase letter, and all following characters must be a dash, lowercase letter, or digit, except the last character, which cannot be a dash"))
		}

	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.DiskName == "" {
		c.DiskName = c.InstanceName
	}

	if c.MachineType == "" {
		c.MachineType = "n1-standard-1"
	}

	if c.RawStateTimeout == "" {
		c.RawStateTimeout = "5m"
	}

	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = "root"
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Process required parameters.
	if c.ProjectId == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a project_id must be specified"))
	}

	if c.SourceImage == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a source_image must be specified"))
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a zone must be specified"))
	}
	if c.Region == "" && len(c.Zone) > 2 {
		// get region from Zone
		region := c.Zone[:len(c.Zone)-2]
		c.Region = region
	}

	stateTimeout, err := time.ParseDuration(c.RawStateTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing state_timeout: %s", err))
	}
	c.stateTimeout = stateTimeout

	if c.AccountFile != "" {
		if err := processAccountFile(&c.account, c.AccountFile); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if c.OmitExternalIP && c.Address != "" {
		errs = packer.MultiErrorAppend(fmt.Errorf("you can not specify an external address when 'omit_external_ip' is true"))
	}

	if c.OmitExternalIP && !c.UseInternalIP {
		errs = packer.MultiErrorAppend(fmt.Errorf("'use_internal_ip' must be true if 'omit_external_ip' is true"))
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}
