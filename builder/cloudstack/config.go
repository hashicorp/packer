//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package cloudstack

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

// Config holds all the details needed to configure the builder.
type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	commonsteps.HTTPConfig `mapstructure:",squash"`
	Comm                   communicator.Config `mapstructure:",squash"`

	// The CloudStack API endpoint we will connect to. It can
	// also be specified via environment variable CLOUDSTACK_API_URL, if set.
	APIURL string `mapstructure:"api_url" required:"true"`
	// The API key used to sign all API requests. It can also
	// be specified via environment variable CLOUDSTACK_API_KEY, if set.
	APIKey string `mapstructure:"api_key" required:"true"`
	// The secret key used to sign all API requests. It
	// can also be specified via environment variable CLOUDSTACK_SECRET_KEY, if
	// set.
	SecretKey string `mapstructure:"secret_key" required:"true"`
	// The time duration to wait for async calls to
	// finish. Defaults to 30m.
	AsyncTimeout time.Duration `mapstructure:"async_timeout" required:"false"`
	// Some cloud providers only allow HTTP GET calls
	// to their CloudStack API. If using such a provider, you need to set this to
	// true in order for the provider to only make GET calls and no POST calls.
	HTTPGetOnly bool `mapstructure:"http_get_only" required:"false"`
	// Set to true to skip SSL verification.
	// Defaults to false.
	SSLNoVerify bool `mapstructure:"ssl_no_verify" required:"false"`
	// List of CIDR's that will have access to the new
	// instance. This is needed in order for any provisioners to be able to
	// connect to the instance. Defaults to [ "0.0.0.0/0" ]. Only required when
	// use_local_ip_address is false.
	CIDRList []string `mapstructure:"cidr_list" required:"false"`
	// If true a temporary security group
	// will be created which allows traffic towards the instance from the
	// cidr_list. This option will be ignored if security_groups is also
	// defined. Requires expunge set to true. Defaults to false.
	CreateSecurityGroup bool `mapstructure:"create_security_group" required:"false"`
	// The name or ID of the disk offering used for the
	// instance. This option is only available (and also required) when using
	// source_iso.
	DiskOffering string `mapstructure:"disk_offering" required:"false"`
	// The size (in GB) of the root disk of the new
	// instance. This option is only available when using source_template.
	DiskSize int64 `mapstructure:"disk_size" required:"false"`
	// If `true` make a call to the CloudStack API, after loading image to
	// cache, requesting to check and detach ISO file (if any) currently
	// attached to a virtual machine. Defaults to `false`. This option is only
	// available when using `source_iso`.
	EjectISO bool `mapstructure:"eject_iso"`
	// Configure the duration time to wait, making sure virtual machine is able
	// to finish installing OS before it ejects safely. Requires `eject_iso`
	// set to `true` and this option is only available when using `source_iso`.
	EjectISODelay time.Duration `mapstructure:"eject_iso_delay"`
	// Set to true to expunge the instance when it is
	// destroyed. Defaults to false.
	Expunge bool `mapstructure:"expunge" required:"false"`
	// The target hypervisor (e.g. XenServer, KVM) for
	// the new template. This option is required when using source_iso.
	Hypervisor string `mapstructure:"hypervisor" required:"false"`
	// The name of the instance. Defaults to
	// "packer-UUID" where UUID is dynamically generated.
	InstanceName string `mapstructure:"instance_name" required:"false"`
	// The display name of the instance. Defaults to "Created by Packer".
	InstanceDisplayName string `mapstructure:"instance_display_name" required:"false"`
	// The name or ID of the network to connect the instance
	// to.
	Network string `mapstructure:"network" required:"true"`
	// The name or ID of the project to deploy the instance
	// to.
	Project string `mapstructure:"project" required:"false"`
	// The public IP address or it's ID used for
	// connecting any provisioners to. If not provided, a temporary public IP
	// address will be associated and released during the Packer run.
	PublicIPAddress string `mapstructure:"public_ip_address" required:"false"`
	// The fixed port you want to configure in the port
	// forwarding rule. Set this attribute if you do not want to use the a random
	// public port.
	PublicPort int `mapstructure:"public_port" required:"false"`
	// A list of security group IDs or
	// names to associate the instance with.
	SecurityGroups []string `mapstructure:"security_groups" required:"false"`
	// The name or ID of the service offering used
	// for the instance.
	ServiceOffering string `mapstructure:"service_offering" required:"true"`
	// Set to true to prevent network
	// ACLs or firewall rules creation. Defaults to false.
	PreventFirewallChanges bool `mapstructure:"prevent_firewall_changes" required:"false"`
	// The name or ID of an ISO that will be mounted
	// before booting the instance. This option is mutually exclusive with
	// source_template. When using source_iso, both disk_offering and
	// hypervisor are required.
	SourceISO string `mapstructure:"source_iso" required:"true"`
	// The name or ID of the template used as base
	// template for the instance. This option is mutually exclusive with
	// source_iso.
	SourceTemplate string `mapstructure:"source_template" required:"true"`
	// The name of the temporary SSH key pair
	// to generate. By default, Packer generates a name that looks like
	// `packer_<UUID>`, where `<UUID>` is a 36 character unique identifier.
	TemporaryKeypairName string `mapstructure:"temporary_keypair_name" required:"false"`
	// Set to true to indicate that the
	// provisioners should connect to the local IP address of the instance.
	UseLocalIPAddress bool `mapstructure:"use_local_ip_address" required:"false"`
	// User data to launch with the instance. This is a
	// template engine; see "User Data" bellow for
	// more details. Packer will not automatically wait for a user script to
	// finish before shutting down the instance this must be handled in a
	// provisioner.
	UserData string `mapstructure:"user_data" required:"false"`
	// Path to a file that will be used for the user
	// data when launching the instance. This file will be parsed as a template
	// engine see User Data bellow for more
	// details.
	UserDataFile string `mapstructure:"user_data_file" required:"false"`
	// The name or ID of the zone where the instance will be
	// created.
	Zone string `mapstructure:"zone" required:"true"`
	// The name of the new template. Defaults to
	// `packer-{{timestamp}}` where timestamp will be the current time.
	TemplateName string `mapstructure:"template_name" required:"false"`
	// The display text of the new template.
	// Defaults to the template_name.
	TemplateDisplayText string `mapstructure:"template_display_text" required:"false"`
	// The name or ID of the template OS for the new
	// template that will be created.
	TemplateOS string `mapstructure:"template_os" required:"true"`
	// Set to true to indicate that the template
	// is featured. Defaults to false.
	TemplateFeatured bool `mapstructure:"template_featured" required:"false"`
	// Set to true to indicate that the template
	// is available for all accounts. Defaults to false.
	TemplatePublic bool `mapstructure:"template_public" required:"false"`
	// Set to true to indicate the
	// template should be password enabled. Defaults to false.
	TemplatePasswordEnabled bool `mapstructure:"template_password_enabled" required:"false"`
	// Set to true to indicate the template
	// requires hardware-assisted virtualization. Defaults to false.
	TemplateRequiresHVM bool `mapstructure:"template_requires_hvm" required:"false"`
	// Set to true to indicate that the template
	// contains tools to support dynamic scaling of VM cpu/memory. Defaults to
	// false.
	TemplateScalable bool `mapstructure:"template_scalable" required:"false"`
	//
	TemplateTag string `mapstructure:"template_tag"`

	Tags map[string]string `mapstructure:"tags"`

	ctx interpolate.Context
}

// NewConfig parses and validates the given config.
func (c *Config) Prepare(raws ...interface{}) error {
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"user_data",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError

	// Set some defaults.
	if c.APIURL == "" {
		// Default to environment variable for api_url, if it exists
		c.APIURL = os.Getenv("CLOUDSTACK_API_URL")
	}

	if c.APIKey == "" {
		// Default to environment variable for api_key, if it exists
		c.APIKey = os.Getenv("CLOUDSTACK_API_KEY")
	}

	if c.SecretKey == "" {
		// Default to environment variable for secret_key, if it exists
		c.SecretKey = os.Getenv("CLOUDSTACK_SECRET_KEY")
	}

	if c.AsyncTimeout == 0 {
		c.AsyncTimeout = 30 * time.Minute
	}

	if len(c.CIDRList) == 0 {
		c.CIDRList = []string{"0.0.0.0/0"}
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.InstanceDisplayName == "" {
		c.InstanceDisplayName = "Created by Packer"
	}

	if c.TemplateName == "" {
		name, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs,
				fmt.Errorf("Unable to parse template name: %s ", err))
		}

		c.TemplateName = name
	}

	if c.TemplateDisplayText == "" {
		c.TemplateDisplayText = c.TemplateName
	}

	// If we are not given an explicit keypair, ssh_password or ssh_private_key_file,
	// then create a temporary one, but only if the temporary_keypair_name has not
	// been provided.
	if c.Comm.SSHKeyPairName == "" && c.Comm.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" {
		c.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	// Process required parameters.
	if c.APIURL == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a api_url must be specified"))
	}

	if c.APIKey == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a api_key must be specified"))
	}

	if c.SecretKey == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a secret_key must be specified"))
	}

	if c.Network == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a network must be specified"))
	}

	if c.CreateSecurityGroup && !c.Expunge {
		errs = packersdk.MultiErrorAppend(errs, errors.New("auto creating a temporary security group requires expunge"))
	}

	if c.ServiceOffering == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a service_offering must be specified"))
	}

	if c.SourceISO == "" && c.SourceTemplate == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("either source_iso or source_template must be specified"))
	}

	if c.SourceISO != "" && c.SourceTemplate != "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("only one of source_iso or source_template can be specified"))
	}

	if c.SourceISO != "" && c.DiskOffering == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("a disk_offering must be specified when using source_iso"))
	}

	if c.SourceISO != "" && c.Hypervisor == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("a hypervisor must be specified when using source_iso"))
	}

	if c.TemplateOS == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a template_os must be specified"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("only one of user_data or user_data_file can be specified"))
	}

	if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if c.Zone == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("a zone must be specified"))
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	// Check for errors and return if we have any.
	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}
