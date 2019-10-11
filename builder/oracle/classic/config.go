package classic

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
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
	PVConfig            `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`
	attribs             map[string]interface{}

	// Access config overrides
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	IdentityDomain string `mapstructure:"identity_domain"`
	APIEndpoint    string `mapstructure:"api_endpoint"`
	apiEndpointURL *url.URL

	// Image
	ImageName            string        `mapstructure:"image_name"`
	Shape                string        `mapstructure:"shape"`
	SourceImageList      string        `mapstructure:"source_image_list"`
	SourceImageListEntry int           `mapstructure:"source_image_list_entry"`
	SnapshotTimeout      time.Duration `mapstructure:"snapshot_timeout"`
	DestImageList        string        `mapstructure:"dest_image_list"`
	// Attributes and Attributes file are both optional and mutually exclusive.
	Attributes     string `mapstructure:"attributes"`
	AttributesFile string `mapstructure:"attributes_file"`
	// Optional; if you don't enter anything, the image list description
	// will read "Packer-built image list"
	DestImageListDescription string `mapstructure:"image_description"`
	// Optional. Describes what computers are allowed to reach your instance
	// via SSH. This whitelist must contain the computer you're running Packer
	// from. It defaults to public-internet, meaning that you can SSH into your
	// instance from anywhere as long as you have the right keys
	SSHSourceList string `mapstructure:"ssh_source_list"`

	ctx interpolate.Context
}

func (c *Config) Identifier(s string) string {
	return fmt.Sprintf("/Compute-%s/%s/%s", c.IdentityDomain, c.Username, s)
}

func NewConfig(raws ...interface{}) (*Config, error) {
	c := &Config{}

	// Decode from template
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, fmt.Errorf("Failed to mapstructure Config: %+v", err)
	}

	c.apiEndpointURL, err = url.Parse(c.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Error parsing API Endpoint: %s", err)
	}
	// Set default source list
	if c.SSHSourceList == "" {
		c.SSHSourceList = "seciplist:/oracle/public/public-internet"
	}

	if c.SnapshotTimeout == 0 {
		c.SnapshotTimeout = 20 * time.Minute
	}

	// Validate that all required fields are present
	var errs *packer.MultiError
	required := map[string]string{
		"username":          c.Username,
		"password":          c.Password,
		"api_endpoint":      c.APIEndpoint,
		"identity_domain":   c.IdentityDomain,
		"source_image_list": c.SourceImageList,
		"dest_image_list":   c.DestImageList,
		"shape":             c.Shape,
	}
	for k, v := range required {
		if v == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("You must specify a %s.", k))
		}
	}

	// Object names can contain only alphanumeric characters, hyphens, underscores, and periods
	reValidObject := regexp.MustCompile("^[a-zA-Z0-9-._/]+$")
	var objectValidation = []struct {
		name  string
		value string
	}{
		{"dest_image_list", c.DestImageList},
		{"image_name", c.ImageName},
	}
	for _, ov := range objectValidation {
		if !reValidObject.MatchString(ov.value) {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("%s can contain only alphanumeric characters, hyphens, underscores, and periods.", ov.name))
		}
	}

	if c.Attributes != "" && c.AttributesFile != "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Only one of user_data or user_data_file can be specified."))
	} else if c.AttributesFile != "" {
		if _, err := os.Stat(c.AttributesFile); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("attributes_file not found: %s", c.AttributesFile))
		}
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	// unpack attributes from json into config
	var data map[string]interface{}

	if c.Attributes != "" {
		err := json.Unmarshal([]byte(c.Attributes), &data)
		if err != nil {
			err = fmt.Errorf("Problem parsing json from attributes: %s", err)
			packer.MultiErrorAppend(errs, err)
		}
		c.attribs = data
	} else if c.AttributesFile != "" {
		fidata, err := ioutil.ReadFile(c.AttributesFile)
		if err != nil {
			err = fmt.Errorf("Problem reading attributes_file: %s", err)
			packer.MultiErrorAppend(errs, err)
		}
		err = json.Unmarshal(fidata, &data)
		c.attribs = data
		if err != nil {
			err = fmt.Errorf("Problem parsing json from attributes_file: %s", err)
			packer.MultiErrorAppend(errs, err)
		}
		c.attribs = data
	}

	return c, nil
}
