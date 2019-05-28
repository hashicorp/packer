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

type AmiFilterOptions struct {
	Filters    map[*string]*string
	Owners     []*string
	MostRecent bool `mapstructure:"most_recent"`
}

func (d *AmiFilterOptions) Empty() bool {
	return len(d.Owners) == 0 && len(d.Filters) == 0
}

func (d *AmiFilterOptions) NoOwner() bool {
	return len(d.Owners) == 0
}

type SubnetFilterOptions struct {
	Filters  map[*string]*string
	MostFree bool `mapstructure:"most_free"`
	Random   bool `mapstructure:"random"`
}

func (d *SubnetFilterOptions) Empty() bool {
	return len(d.Filters) == 0
}

type VpcFilterOptions struct {
	Filters map[*string]*string
}

func (d *VpcFilterOptions) Empty() bool {
	return len(d.Filters) == 0
}

type SecurityGroupFilterOptions struct {
	Filters map[*string]*string
}

func (d *SecurityGroupFilterOptions) Empty() bool {
	return len(d.Filters) == 0
}

// RunConfig contains configuration for running an instance from a source
// AMI and details on how to access that launched image.
type RunConfig struct {
	// If using a non-default VPC,
    // public IP addresses are not provided by default. If this is true, your
    // new instance will get a Public IP. default: false
	AssociatePublicIpAddress          bool                       `mapstructure:"associate_public_ip_address" required:"false"`
	// Destination availability zone to launch
    // instance in. Leave this empty to allow Amazon to auto-assign.
	AvailabilityZone                  string                     `mapstructure:"availability_zone" required:"false"`
	// Requires spot_price to be set. The
    // required duration for the Spot Instances (also known as Spot blocks). This
    // value must be a multiple of 60 (60, 120, 180, 240, 300, or 360). You can't
    // specify an Availability Zone group or a launch group if you specify a
    // duration.
	BlockDurationMinutes              int64                      `mapstructure:"block_duration_minutes" required:"false"`
	// Packer normally stops the build
    // instance after all provisioners have run. For Windows instances, it is
    // sometimes desirable to run
    // Sysprep
    // which will stop the instance for you. If this is set to true, Packer
    // will not stop the instance but will assume that you will send the stop
    // signal yourself through your final provisioner. You can do this with a
    // windows-shell
    // provisioner.
	DisableStopInstance               bool                       `mapstructure:"disable_stop_instance" required:"false"`
	// Mark instance as EBS
    // Optimized.
    // Default false.
	EbsOptimized                      bool                       `mapstructure:"ebs_optimized" required:"false"`
	// Enabling T2 Unlimited allows the source
    // instance to burst additional CPU beyond its available CPU
    // Credits
    // for as long as the demand exists. This is in contrast to the standard
    // configuration that only allows an instance to consume up to its available
    // CPU Credits. See the AWS documentation for T2
    // Unlimited
    // and the T2 Unlimited Pricing section of the Amazon EC2 On-Demand
    // Pricing document for more
    // information. By default this option is disabled and Packer will set up a
    // T2
    // Standard
    // instance instead.
	EnableT2Unlimited                 bool                       `mapstructure:"enable_t2_unlimited" required:"false"`
	// The name of an IAM instance
    // profile
    // to launch the EC2 instance with.
	IamInstanceProfile                string                     `mapstructure:"iam_instance_profile" required:"false"`
	// Automatically terminate instances on
    // shutdown in case Packer exits ungracefully. Possible values are stop and
    // terminate. Defaults to stop.
	InstanceInitiatedShutdownBehavior string                     `mapstructure:"shutdown_behavior" required:"false"`
	// The EC2 instance type to use while building the
    // AMI, such as t2.small.
	InstanceType                      string                     `mapstructure:"instance_type" required:"true"`
	// Filters used to populate the
    // security_group_ids field. Example:
	SecurityGroupFilter               SecurityGroupFilterOptions `mapstructure:"security_group_filter" required:"false"`
	// Tags to apply to the instance
    // that is launched to create the AMI. These tags are not applied to the
    // resulting AMI unless they're duplicated in tags. This is a template
    // engine, see Build template
    // data for more information.
	RunTags                           map[string]string          `mapstructure:"run_tags" required:"false"`
	// The ID (not the name) of the security
    // group to assign to the instance. By default this is not set and Packer will
    // automatically create a new temporary security group to allow SSH access.
    // Note that if this is specified, you must be sure the security group allows
    // access to the ssh_port given below.
	SecurityGroupId                   string                     `mapstructure:"security_group_id" required:"false"`
	// A list of security groups as
    // described above. Note that if this is specified, you must omit the
    // security_group_id.
	SecurityGroupIds                  []string                   `mapstructure:"security_group_ids" required:"false"`
	// The source AMI whose root volume will be copied and
    // provisioned on the currently running instance. This must be an EBS-backed
    // AMI with a root volume snapshot that you have access to. Note: this is not
    // used when from_scratch is set to true.
	SourceAmi                         string                     `mapstructure:"source_ami" required:"true"`
	// Filters used to populate the source_ami
    // field. Example:
	SourceAmiFilter                   AmiFilterOptions           `mapstructure:"source_ami_filter" required:"false"`
	// a list of acceptable instance
    // types to run your build on. We will request a spot instance using the max
    // price of spot_price and the allocation strategy of "lowest price".
    // Your instance will be launched on an instance type of the lowest available
    // price that you have in your list.  This is used in place of instance_type.
    // You may only set either spot_instance_types or instance_type, not both.
    // This feature exists to help prevent situations where a Packer build fails
    // because a particular availability zone does not have capacity for the
    // specific instance_type requested in instance_type.
	SpotInstanceTypes                 []string                   `mapstructure:"spot_instance_types" required:"false"`
	// The maximum hourly price to pay for a spot instance
    // to create the AMI. Spot instances are a type of instance that EC2 starts
    // when the current spot price is less than the maximum price you specify.
    // Spot price will be updated based on available spot instance capacity and
    // current spot instance requests. It may save you some costs. You can set
    // this to auto for Packer to automatically discover the best spot price or
    // to "0" to use an on demand instance (default).
	SpotPrice                         string                     `mapstructure:"spot_price" required:"false"`
	// Required if spot_price is set to
    // auto. This tells Packer what sort of AMI you're launching to find the
    // best spot price. This must be one of: Linux/UNIX, SUSE Linux,
    // Windows, Linux/UNIX (Amazon VPC), SUSE Linux (Amazon VPC),
    // Windows (Amazon VPC)
	SpotPriceAutoProduct              string                     `mapstructure:"spot_price_auto_product" required:"false"`
	// Requires spot_price to be
    // set. This tells Packer to apply tags to the spot request that is issued.
	SpotTags                          map[string]string          `mapstructure:"spot_tags" required:"false"`
	// Filters used to populate the subnet_id field.
    // Example:
	SubnetFilter                      SubnetFilterOptions        `mapstructure:"subnet_filter" required:"false"`
	// If using VPC, the ID of the subnet, such as
    // subnet-12345def, where Packer will launch the EC2 instance. This field is
    // required if you are using an non-default VPC.
	SubnetId                          string                     `mapstructure:"subnet_id" required:"false"`
	// The name of the temporary key pair to
    // generate. By default, Packer generates a name that looks like
    // packer_<UUID>, where <UUID> is a 36 character unique identifier.
	TemporaryKeyPairName              string                     `mapstructure:"temporary_key_pair_name" required:"false"`
	// A list of IPv4
    // CIDR blocks to be authorized access to the instance, when packer is creating a temporary security group.
	TemporarySGSourceCidrs            []string                   `mapstructure:"temporary_security_group_source_cidrs" required:"false"`
	// User data to apply when launching the instance. Note
    // that you need to be careful about escaping characters due to the templates
    // being JSON. It is often more convenient to use user_data_file, instead.
    // Packer will not automatically wait for a user script to finish before
    // shutting down the instance this must be handled in a provisioner.
	UserData                          string                     `mapstructure:"user_data" required:"false"`
	// Path to a file that will be used for the user
    // data when launching the instance.
	UserDataFile                      string                     `mapstructure:"user_data_file" required:"false"`
	// Filters used to populate the vpc_id field.
    // vpc_id take precedence over this.
    // Example:
	VpcFilter                         VpcFilterOptions           `mapstructure:"vpc_filter" required:"false"`
	// If launching into a VPC subnet, Packer needs the VPC ID
    // in order to create a temporary security group within the VPC. Requires
    // subnet_id to be set. If this field is left blank, Packer will try to get
    // the VPC ID from the subnet_id.
	VpcId                             string                     `mapstructure:"vpc_id" required:"false"`
	// The timeout for waiting for a Windows
    // password for Windows instances. Defaults to 20 minutes. Example value:
    // 10m
	WindowsPasswordTimeout            time.Duration              `mapstructure:"windows_password_timeout" required:"false"`

	// Communicator settings
	Comm communicator.Config `mapstructure:",squash"`
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
	if c.Comm.SSHInterface != "public_ip" &&
		c.Comm.SSHInterface != "private_ip" &&
		c.Comm.SSHInterface != "public_dns" &&
		c.Comm.SSHInterface != "private_dns" &&
		c.Comm.SSHInterface != "" {
		errs = append(errs, fmt.Errorf("Unknown interface type: %s", c.Comm.SSHInterface))
	}

	if c.Comm.SSHKeyPairName != "" {
		if c.Comm.Type == "winrm" && c.Comm.WinRMPassword == "" && c.Comm.SSHPrivateKeyFile == "" {
			errs = append(errs, fmt.Errorf("ssh_private_key_file must be provided to retrieve the winrm password when using ssh_keypair_name."))
		} else if c.Comm.SSHPrivateKeyFile == "" && !c.Comm.SSHAgentAuth {
			errs = append(errs, fmt.Errorf("ssh_private_key_file must be provided or ssh_agent_auth enabled when ssh_keypair_name is specified."))
		}
	}

	if c.SourceAmi == "" && c.SourceAmiFilter.Empty() {
		errs = append(errs, fmt.Errorf("A source_ami or source_ami_filter must be specified"))
	}

	if c.SourceAmi == "" && c.SourceAmiFilter.NoOwner() {
		errs = append(errs, fmt.Errorf("For security reasons, your source AMI filter must declare an owner."))
	}

	if c.InstanceType == "" && len(c.SpotInstanceTypes) == 0 {
		errs = append(errs, fmt.Errorf("either instance_type or "+
			"spot_instance_types must be specified"))
	}

	if c.InstanceType != "" && len(c.SpotInstanceTypes) > 0 {
		errs = append(errs, fmt.Errorf("either instance_type or "+
			"spot_instance_types must be specified, not both"))
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
				"spot_tags should not be set when not requesting a spot instance"))
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

	if len(c.TemporarySGSourceCidrs) == 0 {
		c.TemporarySGSourceCidrs = []string{"0.0.0.0/0"}
	} else {
		for _, cidr := range c.TemporarySGSourceCidrs {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				errs = append(errs, fmt.Errorf("Error parsing CIDR in temporary_security_group_source_cidrs: %s", err.Error()))
			}
		}
	}

	if c.InstanceInitiatedShutdownBehavior == "" {
		c.InstanceInitiatedShutdownBehavior = "stop"
	} else if !reShutdownBehavior.MatchString(c.InstanceInitiatedShutdownBehavior) {
		errs = append(errs, fmt.Errorf("shutdown_behavior only accepts 'stop' or 'terminate' values."))
	}

	if c.EnableT2Unlimited {
		if c.SpotPrice != "" {
			errs = append(errs, fmt.Errorf("Error: T2 Unlimited cannot be used in conjuction with Spot Instances"))
		}
		firstDotIndex := strings.Index(c.InstanceType, ".")
		if firstDotIndex == -1 {
			errs = append(errs, fmt.Errorf("Error determining main Instance Type from: %s", c.InstanceType))
		} else if c.InstanceType[0:firstDotIndex] != "t2" {
			errs = append(errs, fmt.Errorf("Error: T2 Unlimited enabled with a non-T2 Instance Type: %s", c.InstanceType))
		}
	}

	return errs
}

func (c *RunConfig) IsSpotInstance() bool {
	return c.SpotPrice != "" && c.SpotPrice != "0"
}
