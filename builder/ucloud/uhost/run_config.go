package uhost

import (
	"fmt"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
	"regexp"
)

type RunConfig struct {
	Zone            string `mapstructure:"availability_zone"`
	SourceImageId   string `mapstructure:"source_image_id"`
	InstanceType    string `mapstructure:"instance_type"`
	InstanceName    string `mapstructure:"instance_name"`
	BootDiskType    string `mapstructure:"boot_disk_type"`
	VPCId           string `mapstructure:"vpc_id"`
	SubnetId        string `mapstructure:"subnet_id"`
	SecurityGroupId string `mapstructure:"security_group_id"`

	// Communicator settings
	Comm            communicator.Config `mapstructure:",squash"`
	UseSSHPrivateIp bool                `mapstructure:"use_ssh_private_ip"`
}

var instanceNamePattern = regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{1,63}$`)

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	errs := c.Comm.Prepare(ctx)

	if c.Zone == "" {
		errs = append(errs, fmt.Errorf("%q must be set", "availability_zone"))
	}

	if c.SourceImageId == "" {
		errs = append(errs, fmt.Errorf("%q must be set", "source_image_id"))
	}

	if c.InstanceType == "" {
		errs = append(errs, fmt.Errorf("%q must be set", "instance_type"))
	} else if _, err := parseInstanceType(c.InstanceType); err != nil {
		errs = append(errs, err)
	}

	if (c.VPCId != "" && c.SubnetId == "") || (c.VPCId == "" && c.SubnetId != "") {
		errs = append(errs, fmt.Errorf("expected both %q and %q to set or not set", "vpc_id", "subnet_id"))
	}

	if c.BootDiskType == "" {
		c.BootDiskType = "cloud_ssd"
	} else if err := checkStringIn(c.BootDiskType,
		[]string{"local_normal", "local_ssd", "cloud_normal", "cloud_ssd"}); err != nil {
		errs = append(errs, err)
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID()[:8])
	} else if !instanceNamePattern.MatchString(c.InstanceName) {
		errs = append(errs, fmt.Errorf("expected %q to be 1-63 characters and only support chinese, english, numbers, '-_.', got %q", "instance_name", c.InstanceName))
	}

	if c.UseSSHPrivateIp == true && c.VPCId == "" {
		errs = append(errs, fmt.Errorf("%q must be set when use_ssh_private_ip is true", "vpc_id"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
