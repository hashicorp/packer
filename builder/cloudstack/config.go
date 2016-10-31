package cloudstack

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// Config holds all the details needed to configure the builder.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	APIURL       string        `mapstructure:"api_url"`
	APIKey       string        `mapstructure:"api_key"`
	SecretKey    string        `mapstructure:"secret_key"`
	AsyncTimeout time.Duration `mapstructure:"async_timeout"`
	HTTPGetOnly  bool          `mapstructure:"http_get_only"`
	SSLNoVerify  bool          `mapstructure:"ssl_no_verify"`

	DiskOffering      string   `mapstructure:"disk_offering"`
	DiskSize          int64    `mapstructure:"disk_size"`
	CIDRList          []string `mapstructure:"cidr_list"`
	Hypervisor        string   `mapstructure:"hypervisor"`
	InstanceName      string   `mapstructure:"instance_name"`
	Keypair           string   `mapstructure:"keypair"`
	Network           string   `mapstructure:"network"`
	Project           string   `mapstructure:"project"`
	PublicIPAddress   string   `mapstructure:"public_ip_address"`
	ServiceOffering   string   `mapstructure:"service_offering"`
	SourceTemplate    string   `mapstructure:"source_template"`
	SourceISO         string   `mapstructure:"source_iso"`
	UserData          string   `mapstructure:"user_data"`
	UserDataFile      string   `mapstructure:"user_data_file"`
	UseLocalIPAddress bool     `mapstructure:"use_local_ip_address"`
	Zone              string   `mapstructure:"zone"`

	TemplateName            string `mapstructure:"template_name"`
	TemplateDisplayText     string `mapstructure:"template_display_text"`
	TemplateOS              string `mapstructure:"template_os"`
	TemplateFeatured        bool   `mapstructure:"template_featured"`
	TemplatePublic          bool   `mapstructure:"template_public"`
	TemplatePasswordEnabled bool   `mapstructure:"template_password_enabled"`
	TemplateRequiresHVM     bool   `mapstructure:"template_requires_hvm"`
	TemplateScalable        bool   `mapstructure:"template_scalable"`
	TemplateTag             string `mapstructure:"template_tag"`

	ctx            interpolate.Context
	hostAddress    string // The host address used by the communicators.
	instanceSource string // This can be either a template ID or an ISO ID.
}

// NewConfig parses and validates the given config.
func NewConfig(raws ...interface{}) (*Config, error) {
	c := new(Config)
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, err
	}

	var errs *packer.MultiError

	// Set some defaults.
	if c.AsyncTimeout == 0 {
		c.AsyncTimeout = 30 * time.Minute
	}

	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = "root"
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.TemplateName == "" {
		name, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Unable to parse template name: %s ", err))
		}

		c.TemplateName = name
	}

	if c.TemplateDisplayText == "" {
		c.TemplateDisplayText = c.TemplateName
	}

	// Process required parameters.
	if c.APIURL == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a api_url must be specified"))
	}

	if c.APIKey == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a api_key must be specified"))
	}

	if c.SecretKey == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a secret_key must be specified"))
	}

	if len(c.CIDRList) == 0 && !c.UseLocalIPAddress {
		errs = packer.MultiErrorAppend(errs, errors.New("a cidr_list must be specified"))
	}

	if c.Network == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a network must be specified"))
	}

	if c.ServiceOffering == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a service_offering must be specified"))
	}

	if c.SourceISO == "" && c.SourceTemplate == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("either source_iso or source_template must be specified"))
	}

	if c.SourceISO != "" && c.SourceTemplate != "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("only one of source_iso or source_template can be specified"))
	}

	if c.SourceISO != "" && c.DiskOffering == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a disk_offering must be specified when using source_iso"))
	}

	if c.SourceISO != "" && c.Hypervisor == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a hypervisor must be specified when using source_iso"))
	}

	if c.TemplateOS == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a template_os must be specified"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("only one of user_data or user_data_file can be specified"))
	}

	if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a zone must be specified"))
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Check for errors and return if we have any.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return c, nil
}
