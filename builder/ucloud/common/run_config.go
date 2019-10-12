package common

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
	} else if _, err := ParseInstanceType(c.InstanceType); err != nil {
		errs = append(errs, err)
	}

	if (c.VPCId != "" && c.SubnetId == "") || (c.VPCId == "" && c.SubnetId != "") {
		errs = append(errs, fmt.Errorf("expected both %q and %q to set or not set", "vpc_id", "subnet_id"))
	}

	if c.BootDiskType == "" {
		c.BootDiskType = "cloud_ssd"
	} else if err := CheckStringIn(c.BootDiskType,
		[]string{"local_normal", "local_ssd", "cloud_ssd"}); err != nil {
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

	if c.Comm.SSHPassword != "" && len(validateInstancePassword(c.Comm.SSHPassword)) != 0 {
		for _, v := range validateInstancePassword(c.Comm.SSHPassword) {
			errs = append(errs, v)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

var instancePasswordUpperPattern = regexp.MustCompile(`[A-Z]`)
var instancePasswordLowerPattern = regexp.MustCompile(`[a-z]`)
var instancePasswordNumPattern = regexp.MustCompile(`[0-9]`)
var instancePasswordSpecialPattern = regexp.MustCompile(`[` + "`" + `()~!@#$%^&*-+=_|{}\[\]:;'<>,.?/]`)
var instancePasswordPattern = regexp.MustCompile(`^[A-Za-z0-9` + "`" + `()~!@#$%^&*-+=_|{}\[\]:;'<>,.?/]{8,30}$`)

func validateInstancePassword(password string) (errors []error) {
	if !instancePasswordPattern.MatchString(password) {
		errors = append(errors, fmt.Errorf("%q is invalid, should have between 8-30 characters and any characters must be legal, got %q", "ssh_password", password))
	}

	categoryCount := 0
	if instancePasswordUpperPattern.MatchString(password) {
		categoryCount++
	}

	if instancePasswordLowerPattern.MatchString(password) {
		categoryCount++
	}

	if instancePasswordNumPattern.MatchString(password) {
		categoryCount++
	}

	if instancePasswordSpecialPattern.MatchString(password) {
		categoryCount++
	}

	if categoryCount < 2 {
		errors = append(errors, fmt.Errorf("%q is invalid, should have least 2 items of capital letters, lower case letters, numbers and special characters, got %q", "ssh_password", password))
	}

	return
}
