//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package digitalocean

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`
	// The client TOKEN to use to access your account. It
	// can also be specified via environment variable DIGITALOCEAN_API_TOKEN, if
	// set.
	APIToken string `mapstructure:"api_token" required:"true"`
	// Non standard api endpoint URL. Set this if you are
	// using a DigitalOcean API compatible service. It can also be specified via
	// environment variable DIGITALOCEAN_API_URL.
	APIURL string `mapstructure:"api_url" required:"false"`
	// The name (or slug) of the region to launch the droplet
	// in. Consequently, this is the region where the snapshot will be available.
	// See
	// https://developers.digitalocean.com/documentation/v2/#list-all-regions
	// for the accepted region names/slugs.
	Region string `mapstructure:"region" required:"true"`
	// The name (or slug) of the droplet size to use. See
	// https://developers.digitalocean.com/documentation/v2/#list-all-sizes
	// for the accepted size names/slugs.
	Size string `mapstructure:"size" required:"true"`
	// The name (or slug) of the base image to use. This is the
	// image that will be used to launch a new droplet and provision it. See
	// https://developers.digitalocean.com/documentation/v2/#list-all-images
	// for details on how to get a list of the accepted image names/slugs.
	Image string `mapstructure:"image" required:"true"`
	// Set to true to enable private networking
	// for the droplet being created. This defaults to false, or not enabled.
	PrivateNetworking bool `mapstructure:"private_networking" required:"false"`
	// Set to true to enable monitoring for the droplet
	// being created. This defaults to false, or not enabled.
	Monitoring bool `mapstructure:"monitoring" required:"false"`
	// Set to true to enable ipv6 for the droplet being
	// created. This defaults to false, or not enabled.
	IPv6 bool `mapstructure:"ipv6" required:"false"`
	// The name of the resulting snapshot that will
	// appear in your account. Defaults to `packer-{{timestamp}}` (see
	// configuration templates for more info).
	SnapshotName string `mapstructure:"snapshot_name" required:"false"`
	// The regions of the resulting
	// snapshot that will appear in your account.
	SnapshotRegions []string `mapstructure:"snapshot_regions" required:"false"`
	// The time to wait, as a duration string, for a
	// droplet to enter a desired state (such as "active") before timing out. The
	// default state timeout is "6m".
	StateTimeout time.Duration `mapstructure:"state_timeout" required:"false"`
	// How long to wait for an image to be published to the shared image
	// gallery before timing out. If your Packer build is failing on the
	// Publishing to Shared Image Gallery step with the error `Original Error:
	// context deadline exceeded`, but the image is present when you check your
	// Azure dashboard, then you probably need to increase this timeout from
	// its default of "60m" (valid time units include `s` for seconds, `m` for
	// minutes, and `h` for hours.)
	SnapshotTimeout time.Duration `mapstructure:"snapshot_timeout" required:"false"`
	// The name assigned to the droplet. DigitalOcean
	// sets the hostname of the machine to this value.
	DropletName string `mapstructure:"droplet_name" required:"false"`
	// User data to launch with the Droplet. Packer will
	// not automatically wait for a user script to finish before shutting down the
	// instance this must be handled in a provisioner.
	UserData string `mapstructure:"user_data" required:"false"`
	// Path to a file that will be used for the user
	// data when launching the Droplet.
	UserDataFile string `mapstructure:"user_data_file" required:"false"`
	// Tags to apply to the droplet when it is created
	Tags []string `mapstructure:"tags" required:"false"`
	// UUID of the VPC which the droplet will be created in. Before using this,
	// private_networking should be enabled.
	VPCUUID string `mapstructure:"vpc_uuid" required:"false"`
	// Wheter the communicators should use private IP or not (public IP in that case).
	// If the droplet is or going to be accessible only from the local network because
	// it is at behind a firewall, then communicators should use the private IP
	// instead of the public IP. Before using this, private_networking should be enabled.
	ConnectWithPrivateIP bool `mapstructure:"connect_with_private_ip" required:"false"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Defaults
	if c.APIToken == "" {
		// Default to environment variable for api_token, if it exists
		c.APIToken = os.Getenv("DIGITALOCEAN_API_TOKEN")
	}
	if c.APIURL == "" {
		c.APIURL = os.Getenv("DIGITALOCEAN_API_URL")
	}
	if c.SnapshotName == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		// Default to packer-{{ unix timestamp (utc) }}
		c.SnapshotName = def
	}

	if c.DropletName == "" {
		// Default to packer-[time-ordered-uuid]
		c.DropletName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.StateTimeout == 0 {
		// Default to 6 minute timeouts waiting for
		// desired state. i.e waiting for droplet to become active
		c.StateTimeout = 6 * time.Minute
	}

	if c.SnapshotTimeout == 0 {
		// Default to 60 minutes timeout, waiting for snapshot action to finish
		c.SnapshotTimeout = 60 * time.Minute
	}

	var errs *packersdk.MultiError

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}
	if c.APIToken == "" {
		// Required configurations that will display errors if not set
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("api_token for auth must be specified"))
	}

	if c.Region == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("region is required"))
	}

	if c.Size == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("size is required"))
	}

	if c.Image == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("image is required"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("only one of user_data or user_data_file can be specified"))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New(fmt.Sprintf("user_data_file not found: %s", c.UserDataFile)))
		}
	}

	if c.Tags == nil {
		c.Tags = make([]string, 0)
	}
	tagRe := regexp.MustCompile("^[[:alnum:]:_-]{1,255}$")

	for _, t := range c.Tags {
		if !tagRe.MatchString(t) {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("invalid tag: %s", t))
		}
	}

	// Check if the PrivateNetworking is enabled by user before use VPC UUID
	if c.VPCUUID != "" {
		if c.PrivateNetworking != true {
			errs = packersdk.MultiErrorAppend(errs, errors.New("private networking should be enabled to use vpc_uuid"))
		}
	}

	// Check if the PrivateNetworking is enabled by user before use ConnectWithPrivateIP
	if c.ConnectWithPrivateIP == true {
		if c.PrivateNetworking != true {
			errs = packersdk.MultiErrorAppend(errs, errors.New("private networking should be enabled to use connect_with_private_ip"))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packersdk.LogSecretFilter.Set(c.APIToken)
	return nil, nil
}
