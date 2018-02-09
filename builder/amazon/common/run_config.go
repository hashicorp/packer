package common

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

var reShutdownBehavior = regexp.MustCompile("^(stop|terminate)$")

type AmiFilterOptions struct {
	Filters    map[*string]*string
	Owners     []*string
	MostRecent bool `mapstructure:"most_recent"`
}

func (d *AmiFilterOptions) Empty() bool {
	return len(d.Owners) == 0 && len(d.Filters) == 0
}

// RunConfig contains configuration for running an instance from a source
// AMI and details on how to access that launched image.
type RunConfig struct {
	AssociatePublicIpAddress          bool              `mapstructure:"associate_public_ip_address"`
	AvailabilityZone                  string            `mapstructure:"availability_zone"`
	EbsOptimized                      bool              `mapstructure:"ebs_optimized"`
	IamInstanceProfile                string            `mapstructure:"iam_instance_profile"`
	InstanceType                      string            `mapstructure:"instance_type"`
	RunTags                           map[string]string `mapstructure:"run_tags"`
	SourceAmi                         string            `mapstructure:"source_ami"`
	SourceAmiFilter                   AmiFilterOptions  `mapstructure:"source_ami_filter"`
	SpotPrice                         string            `mapstructure:"spot_price"`
	SpotPriceAutoProduct              string            `mapstructure:"spot_price_auto_product"`
	DisableStopInstance               bool              `mapstructure:"disable_stop_instance"`
	SecurityGroupId                   string            `mapstructure:"security_group_id"`
	SecurityGroupIds                  []string          `mapstructure:"security_group_ids"`
	TemporarySGSourceCidr             string            `mapstructure:"temporary_security_group_source_cidr"`
	SubnetId                          string            `mapstructure:"subnet_id"`
	TemporaryKeyPairName              string            `mapstructure:"temporary_key_pair_name"`
	UserData                          string            `mapstructure:"user_data"`
	UserDataFile                      string            `mapstructure:"user_data_file"`
	WindowsPasswordTimeout            time.Duration     `mapstructure:"windows_password_timeout"`
	VpcId                             string            `mapstructure:"vpc_id"`
	InstanceInitiatedShutdownBehavior string            `mapstructure:"shutdown_behavior"`

	// Communicator settings
	Comm           communicator.Config `mapstructure:",squash"`
	SSHKeyPairName string              `mapstructure:"ssh_keypair_name"`
	SSHInterface   string              `mapstructure:"ssh_interface"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// If we are not given an explicit ssh_keypair_name or
	// ssh_private_key_file, then create a temporary one, but only if the
	// temporary_key_pair_name has not been provided and we are not using
	// ssh_password.
	if c.SSHKeyPairName == "" && c.TemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKey == "" && c.Comm.SSHPassword == "" {

		c.TemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	if c.WindowsPasswordTimeout == 0 {
		c.WindowsPasswordTimeout = 20 * time.Minute
	}

	if c.RunTags == nil {
		c.RunTags = make(map[string]string)
	}

	// Validation
	errs := c.Comm.Prepare(ctx)

	// Valadating ssh_interface
	if c.SSHInterface != "public_ip" &&
		c.SSHInterface != "private_ip" &&
		c.SSHInterface != "public_dns" &&
		c.SSHInterface != "private_dns" &&
		c.SSHInterface != "" {
		errs = append(errs, errors.New(fmt.Sprintf("Unknown interface type: %s", c.SSHInterface)))
	}

	if c.SSHKeyPairName != "" {
		if c.Comm.Type == "winrm" && c.Comm.WinRMPassword == "" && c.Comm.SSHPrivateKey == "" {
			errs = append(errs, errors.New("ssh_private_key_file must be provided to retrieve the winrm password when using ssh_keypair_name."))
		} else if c.Comm.SSHPrivateKey == "" && !c.Comm.SSHAgentAuth {
			errs = append(errs, errors.New("ssh_private_key_file must be provided or ssh_agent_auth enabled when ssh_keypair_name is specified."))
		}
	}

	if c.SourceAmi == "" && c.SourceAmiFilter.Empty() {
		errs = append(errs, errors.New("A source_ami or source_ami_filter must be specified"))
	}

	if c.InstanceType == "" {
		errs = append(errs, errors.New("An instance_type must be specified"))
	}

	if c.SpotPrice == "auto" {
		if c.SpotPriceAutoProduct == "" {
			errs = append(errs, errors.New(
				"spot_price_auto_product must be specified when spot_price is auto"))
		}
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = append(errs, fmt.Errorf("Only one of user_data or user_data_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = append(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if c.SecurityGroupId != "" {
		if len(c.SecurityGroupIds) > 0 {
			errs = append(errs, fmt.Errorf("Only one of security_group_id or security_group_ids can be specified."))
		} else {
			c.SecurityGroupIds = []string{c.SecurityGroupId}
			c.SecurityGroupId = ""
		}
	}

	if c.TemporarySGSourceCidr == "" {
		c.TemporarySGSourceCidr = "0.0.0.0/0"
	} else {
		if _, _, err := net.ParseCIDR(c.TemporarySGSourceCidr); err != nil {
			errs = append(errs, fmt.Errorf("Error parsing temporary_security_group_source_cidr: %s", err.Error()))
		}
	}

	if c.InstanceInitiatedShutdownBehavior == "" {
		c.InstanceInitiatedShutdownBehavior = "stop"
	} else if !reShutdownBehavior.MatchString(c.InstanceInitiatedShutdownBehavior) {
		errs = append(errs, fmt.Errorf("shutdown_behavior only accepts 'stop' or 'terminate' values."))
	}

	return errs
}

func (c *RunConfig) IsSpotInstance() bool {
	return c.SpotPrice != "" && c.SpotPrice != "0"
}
