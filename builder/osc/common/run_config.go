package common

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

var reShutdownBehavior = regexp.MustCompile("^(stop|terminate)$")

type OmiFilterOptions struct {
	Filters    map[string]string
	Owners     []string
	MostRecent bool `mapstructure:"most_recent"`
}

func (d *OmiFilterOptions) Empty() bool {
	return len(d.Owners) == 0 && len(d.Filters) == 0
}

func (d *OmiFilterOptions) NoOwner() bool {
	return len(d.Owners) == 0
}

type SubnetFilterOptions struct {
	Filters  map[string]string
	MostFree bool `mapstructure:"most_free"`
	Random   bool `mapstructure:"random"`
}

func (d *SubnetFilterOptions) Empty() bool {
	return len(d.Filters) == 0
}

type NetFilterOptions struct {
	Filters map[string]string
}

func (d *NetFilterOptions) Empty() bool {
	return len(d.Filters) == 0
}

type SecurityGroupFilterOptions struct {
	Filters map[string]string
}

func (d *SecurityGroupFilterOptions) Empty() bool {
	return len(d.Filters) == 0
}

// RunConfig contains configuration for running an vm from a source
// AMI and details on how to access that launched image.
type RunConfig struct {
	AssociatePublicIpAddress    bool                       `mapstructure:"associate_public_ip_address"`
	Subregion                   string                     `mapstructure:"subregion_name"`
	BlockDurationMinutes        int64                      `mapstructure:"block_duration_minutes"`
	DisableStopVm               bool                       `mapstructure:"disable_stop_vm"`
	BsuOptimized                bool                       `mapstructure:"bsu_optimized"`
	EnableT2Unlimited           bool                       `mapstructure:"enable_t2_unlimited"`
	IamVmProfile                string                     `mapstructure:"iam_vm_profile"`
	VmInitiatedShutdownBehavior string                     `mapstructure:"shutdown_behavior"`
	VmType                      string                     `mapstructure:"vm_type"`
	SecurityGroupFilter         SecurityGroupFilterOptions `mapstructure:"security_group_filter"`
	RunTags                     map[string]string          `mapstructure:"run_tags"`
	SecurityGroupId             string                     `mapstructure:"security_group_id"`
	SecurityGroupIds            []string                   `mapstructure:"security_group_ids"`
	SourceOmi                   string                     `mapstructure:"source_omi"`
	SourceOmiFilter             OmiFilterOptions           `mapstructure:"source_omi_filter"`
	SpotPrice                   string                     `mapstructure:"spot_price"`
	SpotPriceAutoProduct        string                     `mapstructure:"spot_price_auto_product"`
	SpotTags                    map[string]string          `mapstructure:"spot_tags"`
	SubnetFilter                SubnetFilterOptions        `mapstructure:"subnet_filter"`
	SubnetId                    string                     `mapstructure:"subnet_id"`
	TemporaryKeyPairName        string                     `mapstructure:"temporary_key_pair_name"`
	TemporarySGSourceCidr       string                     `mapstructure:"temporary_security_group_source_cidr"`
	UserData                    string                     `mapstructure:"user_data"`
	UserDataFile                string                     `mapstructure:"user_data_file"`
	NetFilter                   NetFilterOptions           `mapstructure:"net_filter"`
	NetId                       string                     `mapstructure:"net_id"`
	WindowsPasswordTimeout      time.Duration              `mapstructure:"windows_password_timeout"`

	// Communicator settings
	Comm         communicator.Config `mapstructure:",squash"`
	SSHInterface string              `mapstructure:"ssh_interface"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// If we are not given an explicit ssh_keypair_name or
	// ssh_private_key_file, then create a temporary one, but only if the
	// temporary_key_pair_name has not been provided and we are not using
	// ssh_password.
	if c.Comm.SSHKeyPairName == "" && c.Comm.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" {

		c.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	if c.WindowsPasswordTimeout == 0 {
		c.WindowsPasswordTimeout = 20 * time.Minute
	}

	if c.RunTags == nil {
		c.RunTags = make(map[string]string)
	}

	// Validation
	errs := c.Comm.Prepare(ctx)

	// Validating ssh_interface
	if c.SSHInterface != "public_ip" &&
		c.SSHInterface != "private_ip" &&
		c.SSHInterface != "public_dns" &&
		c.SSHInterface != "private_dns" &&
		c.SSHInterface != "" {
		errs = append(errs, fmt.Errorf("Unknown interface type: %s", c.SSHInterface))
	}

	if c.Comm.SSHKeyPairName != "" {
		if c.Comm.Type == "winrm" && c.Comm.WinRMPassword == "" && c.Comm.SSHPrivateKeyFile == "" {
			errs = append(errs, fmt.Errorf("ssh_private_key_file must be provided to retrieve the winrm password when using ssh_keypair_name."))
		} else if c.Comm.SSHPrivateKeyFile == "" && !c.Comm.SSHAgentAuth {
			errs = append(errs, fmt.Errorf("ssh_private_key_file must be provided or ssh_agent_auth enabled when ssh_keypair_name is specified."))
		}
	}

	if c.SourceOmi == "" && c.SourceOmiFilter.Empty() {
		errs = append(errs, fmt.Errorf("A source_omi or source_omi_filter must be specified"))
	}

	if c.SourceOmi == "" && c.SourceOmiFilter.NoOwner() {
		errs = append(errs, fmt.Errorf("For security reasons, your source AMI filter must declare an owner."))
	}

	if c.VmType == "" {
		errs = append(errs, fmt.Errorf("An vm_type must be specified"))
	}

	if c.BlockDurationMinutes%60 != 0 {
		errs = append(errs, fmt.Errorf(
			"block_duration_minutes must be multiple of 60"))
	}

	if c.SpotPrice == "auto" {
		if c.SpotPriceAutoProduct == "" {
			errs = append(errs, fmt.Errorf(
				"spot_price_auto_product must be specified when spot_price is auto"))
		}
	}

	if c.SpotPriceAutoProduct != "" {
		if c.SpotPrice != "auto" {
			errs = append(errs, fmt.Errorf(
				"spot_price should be set to auto when spot_price_auto_product is specified"))
		}
	}

	if c.SpotTags != nil {
		if c.SpotPrice == "" || c.SpotPrice == "0" {
			errs = append(errs, fmt.Errorf(
				"spot_tags should not be set when not requesting a spot vm"))
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

	if c.VmInitiatedShutdownBehavior == "" {
		c.VmInitiatedShutdownBehavior = "stop"
	} else if !reShutdownBehavior.MatchString(c.VmInitiatedShutdownBehavior) {
		errs = append(errs, fmt.Errorf("shutdown_behavior only accepts 'stop' or 'terminate' values."))
	}

	if c.EnableT2Unlimited {
		if c.SpotPrice != "" {
			errs = append(errs, fmt.Errorf("Error: T2 Unlimited cannot be used in conjuction with Spot Vms"))
		}
		firstDotIndex := strings.Index(c.VmType, ".")
		if firstDotIndex == -1 {
			errs = append(errs, fmt.Errorf("Error determining main Vm Type from: %s", c.VmType))
		} else if c.VmType[0:firstDotIndex] != "t2" {
			errs = append(errs, fmt.Errorf("Error: T2 Unlimited enabled with a non-T2 Vm Type: %s", c.VmType))
		}
	}

	return errs
}

func (c *RunConfig) IsSpotVm() bool {
	return c.SpotPrice != "" && c.SpotPrice != "0"
}
