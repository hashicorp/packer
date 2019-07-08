package googlecompute

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"golang.org/x/oauth2/jwt"
	compute "google.golang.org/api/compute/v1"
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

	AcceleratorType              string                         `mapstructure:"accelerator_type"`
	AcceleratorCount             int64                          `mapstructure:"accelerator_count"`
	Address                      string                         `mapstructure:"address"`
	DisableDefaultServiceAccount bool                           `mapstructure:"disable_default_service_account"`
	DiskName                     string                         `mapstructure:"disk_name"`
	DiskSizeGb                   int64                          `mapstructure:"disk_size"`
	DiskType                     string                         `mapstructure:"disk_type"`
	ImageName                    string                         `mapstructure:"image_name"`
	ImageDescription             string                         `mapstructure:"image_description"`
	ImageEncryptionKey           *compute.CustomerEncryptionKey `mapstructure:"image_encryption_key"`
	ImageFamily                  string                         `mapstructure:"image_family"`
	ImageLabels                  map[string]string              `mapstructure:"image_labels"`
	ImageLicenses                []string                       `mapstructure:"image_licenses"`
	InstanceName                 string                         `mapstructure:"instance_name"`
	Labels                       map[string]string              `mapstructure:"labels"`
	MachineType                  string                         `mapstructure:"machine_type"`
	Metadata                     map[string]string              `mapstructure:"metadata"`
	MinCpuPlatform               string                         `mapstructure:"min_cpu_platform"`
	Network                      string                         `mapstructure:"network"`
	NetworkProjectId             string                         `mapstructure:"network_project_id"`
	OmitExternalIP               bool                           `mapstructure:"omit_external_ip"`
	OnHostMaintenance            string                         `mapstructure:"on_host_maintenance"`
	Preemptible                  bool                           `mapstructure:"preemptible"`
	RawStateTimeout              string                         `mapstructure:"state_timeout"`
	Region                       string                         `mapstructure:"region"`
	Scopes                       []string                       `mapstructure:"scopes"`
	ServiceAccountEmail          string                         `mapstructure:"service_account_email"`
	SourceImage                  string                         `mapstructure:"source_image"`
	SourceImageFamily            string                         `mapstructure:"source_image_family"`
	SourceImageProjectId         string                         `mapstructure:"source_image_project_id"`
	StartupScriptFile            string                         `mapstructure:"startup_script_file"`
	Subnetwork                   string                         `mapstructure:"subnetwork"`
	Tags                         []string                       `mapstructure:"tags"`
	UseInternalIP                bool                           `mapstructure:"use_internal_ip"`
	MetadataFiles                map[string]string              `mapstructure:"metadata_files"`
	Zone                         string                         `mapstructure:"zone"`

	Account            *jwt.Config
	stateTimeout       time.Duration
	imageAlreadyExists bool
	ctx                interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	c.ctx.Funcs = TemplateFuncs
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
	if c.Network == "" && c.Subnetwork == "" {
		c.Network = "default"
	}

	if c.NetworkProjectId == "" {
		c.NetworkProjectId = c.ProjectId
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

	if c.OnHostMaintenance == "MIGRATE" && c.Preemptible {
		errs = packer.MultiErrorAppend(errs,
			errors.New("on_host_maintenance must be TERMINATE when using preemptible instances."))
	}
	// Setting OnHostMaintenance Correct Defaults
	//   "MIGRATE" : Possible and default if Preemptible is false
	//   "TERMINATE": Required if Preemptible is true
	if c.Preemptible {
		c.OnHostMaintenance = "TERMINATE"
	} else {
		if c.OnHostMaintenance == "" {
			c.OnHostMaintenance = "MIGRATE"
		}
	}

	// Make sure user sets a valid value for on_host_maintenance option
	if !(c.OnHostMaintenance == "MIGRATE" || c.OnHostMaintenance == "TERMINATE") {
		errs = packer.MultiErrorAppend(errs,
			errors.New("on_host_maintenance must be one of MIGRATE or TERMINATE."))
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

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Process required parameters.
	if c.ProjectId == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a project_id must be specified"))
	}

	if c.Scopes == nil {
		c.Scopes = []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/compute",
			"https://www.googleapis.com/auth/devstorage.full_control",
		}
	}

	if c.SourceImage == "" && c.SourceImageFamily == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a source_image or source_image_family must be specified"))
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

	err = c.CalcTimeout()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if c.AccountFile != "" {
		cfg, err := ProcessAccountFile(c.AccountFile)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
		c.Account = cfg
	}

	if c.OmitExternalIP && c.Address != "" {
		errs = packer.MultiErrorAppend(fmt.Errorf("you can not specify an external address when 'omit_external_ip' is true"))
	}

	if c.OmitExternalIP && !c.UseInternalIP {
		errs = packer.MultiErrorAppend(fmt.Errorf("'use_internal_ip' must be true if 'omit_external_ip' is true"))
	}

	if c.AcceleratorCount > 0 && len(c.AcceleratorType) == 0 {
		errs = packer.MultiErrorAppend(fmt.Errorf("'accelerator_type' must be set when 'accelerator_count' is more than 0"))
	}

	if c.AcceleratorCount > 0 && c.OnHostMaintenance != "TERMINATE" {
		errs = packer.MultiErrorAppend(fmt.Errorf("'on_host_maintenance' must be set to 'TERMINATE' when 'accelerator_count' is more than 0"))
	}

	// If DisableDefaultServiceAccount is provided, don't allow a value for ServiceAccountEmail
	if c.DisableDefaultServiceAccount && c.ServiceAccountEmail != "" {
		errs = packer.MultiErrorAppend(fmt.Errorf("you may not specify a 'service_account_email' when 'disable_default_service_account' is true"))
	}

	if c.StartupScriptFile != "" {
		if _, err := os.Stat(c.StartupScriptFile); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("startup_script_file: %v", err))
		}
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}

func (c *Config) CalcTimeout() error {
	stateTimeout, err := time.ParseDuration(c.RawStateTimeout)
	if err != nil {
		return fmt.Errorf("Failed parsing state_timeout: %s", err)
	}
	c.stateTimeout = stateTimeout
	return nil
}
