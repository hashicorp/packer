//go:generate packer-sdc struct-markdown
package common

import (
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type RunConfig struct {
	// This is the UCloud availability zone where UHost instance is located. such as: `cn-bj2-02`.
	// You may refer to [list of availability_zone](https://docs.ucloud.cn/api/summary/regionlist)
	Zone string `mapstructure:"availability_zone" required:"true"`
	// This is the ID of base image which you want to create your customized images with.
	SourceImageId string `mapstructure:"source_image_id" required:"true"`
	// The type of UHost instance.
	// You may refer to [list of instance type](https://docs.ucloud.cn/compute/terraform/specification/instance)
	InstanceType string `mapstructure:"instance_type" required:"true"`
	// The name of instance, which contains 1-63 characters and only support Chinese,
	// English, numbers, '-', '\_', '.'.
	InstanceName string `mapstructure:"instance_name" required:"false"`
	// The type of boot disk associated to UHost instance.
	// Possible values are: `cloud_ssd` and `cloud_rssd` for cloud boot disk, `local_normal` and `local_ssd`
	// for local boot disk. (Default: `cloud_ssd`). The `cloud_ssd` and `local_ssd` are not fully supported
	// by all regions as boot disk type, please proceed to UCloud console for more details.
	//
	//~> **Note:** It takes around 10 mins for boot disk initialization when `boot_disk_type` is `local_normal` or `local_ssd`.
	BootDiskType string `mapstructure:"boot_disk_type" required:"false"`
	// The ID of VPC linked to the UHost instance. If not defined `vpc_id`, the instance will use the default VPC in the current region.
	VPCId string `mapstructure:"vpc_id" required:"false"`
	// The ID of subnet under the VPC. If `vpc_id` is defined, the `subnet_id` is mandatory required.
	// If `vpc_id` and `subnet_id` are not defined, the instance will use the default subnet in the current region.
	SubnetId string `mapstructure:"subnet_id" required:"false"`
	// The ID of the fire wall associated to UHost instance. If `security_group_id` is not defined,
	// the instance will use the non-recommended web fire wall, and open port include 22, 3389 by default.
	// It is supported by ICMP fire wall protocols.
	// You may refer to [security group_id](https://docs.ucloud.cn/network/firewall/firewall).
	SecurityGroupId string `mapstructure:"security_group_id" required:"false"`
	// Maximum bandwidth to the elastic public network, measured in Mbps (Mega bit per second). (Default: `10`).
	EipBandwidth int `mapstructure:"eip_bandwidth" required:"false"`
	// Elastic IP charge mode. Possible values are: `traffic` as pay by traffic, `bandwidth` as pay by bandwidth,
	// `post_accurate_bandwidth` as post pay mode. (Default: `traffic`).
	// Note currently default `traffic` eip charge mode not not fully support by all `availability_zone`
	// in the `region`, please proceed to [UCloud console](https://console.ucloud.cn/unet/eip/create) for more details.
	// You may refer to [eip introduction](https://docs.ucloud.cn/unet/eip/introduction).
	EipChargeMode string `mapstructure:"eip_charge_mode" required:"false"`
	// User data to apply when launching the instance.
	// Note that you need to be careful about escaping characters due to the templates
	// being JSON. It is often more convenient to use user_data_file, instead.
	// Packer will not automatically wait for a user script to finish before
	// shutting down the instance this must be handled in a provisioner.
	// You may refer to [user_data_document](https://docs.ucloud.cn/uhost/guide/metadata/userdata)
	UserData string `mapstructure:"user_data" required:"false"`
	// Path to a file that will be used for the user data when launching the instance.
	UserDataFile string `mapstructure:"user_data_file" required:"false"`
	// Specifies a minimum CPU platform for the the VM instance. (Default: `Intel/Auto`).
	// You may refer to [min_cpu_platform](https://docs.ucloud.cn/uhost/introduction/uhost/type_new)
	//    - The Intel CPU platform:
	//        - `Intel/Auto` as the Intel CPU platform version will be selected randomly by system;
	//        - `Intel/IvyBridge` as Intel V2, the version of Intel CPU platform selected by system will be `Intel/IvyBridge` and above;
	//        - `Intel/Haswell` as Intel V3,  the version of Intel CPU platform selected by system will be `Intel/Haswell` and above;
	//        - `Intel/Broadwell` as Intel V4, the version of Intel CPU platform selected by system will be `Intel/Broadwell` and above;
	//        - `Intel/Skylake` as Intel V5, the version of Intel CPU platform selected by system will be `Intel/Skylake` and above;
	//        - `Intel/Cascadelake` as Intel V6, the version of Intel CPU platform selected by system will be `Intel/Cascadelake`;
	//    - The AMD CPU platform:
	//        - `Amd/Auto` as the Amd CPU platform version will be selected randomly by system;
	//        - `Amd/Epyc2` as the version of Amd CPU platform selected by system will be `Amd/Epyc2` and above;
	MinCpuPlatform string `mapstructure:"min_cpu_platform" required:"false"`
	// Communicator settings
	Comm communicator.Config `mapstructure:",squash"`
	// If this value is true, packer will connect to the created UHost instance via a private ip
	// instead of allocating an EIP (elastic public ip).(Default: `false`).
	UseSSHPrivateIp bool `mapstructure:"use_ssh_private_ip"`
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
		[]string{"local_normal", "local_ssd", "cloud_ssd", "cloud_rssd"}); err != nil {
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

	if c.UserData != "" && c.UserDataFile != "" {
		errs = append(errs, fmt.Errorf("only one of user_data or user_data_file can be specified"))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = append(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if c.MinCpuPlatform == "" {
		c.MinCpuPlatform = "Intel/Auto"
	} else if err := CheckStringIn(c.MinCpuPlatform,
		[]string{
			"Intel/Auto",
			"Intel/IvyBridge",
			"Intel/Haswell",
			"Intel/Broadwell",
			"Intel/Skylake",
			"Intel/Cascadelake",
			"Amd/Auto",
			"Amd/Epyc2",
		}); err != nil {
		errs = append(errs, err)
	}

	if c.EipChargeMode == "" {
		c.EipChargeMode = "traffic"
	} else if err := CheckStringIn(c.EipChargeMode, []string{"traffic", "bandwidth", "post_accurate_bandwidth"}); err != nil {
		errs = append(errs, err)
	}

	if c.EipBandwidth == 0 {
		c.EipBandwidth = 10
	}

	return errs
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
