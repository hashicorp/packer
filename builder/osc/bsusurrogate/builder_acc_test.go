/*
Deregister the test image with
aws ec2 deregister-image --image-id $(aws ec2 describe-images --output text --filters "Name=name,Values=packer-test-packer-test-dereg" --query 'Images[*].{ID:ImageId}')
*/
package bsusurrogate

import (
	"testing"

	builderT "github.com/hashicorp/packer/helper/builder/testing"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck:             func() { testAccPreCheck(t) },
		Builder:              &Builder{},
		Template:             testBuilderAccBasic,
		SkipArtifactTeardown: true,
	})
}

func testAccPreCheck(t *testing.T) {
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"region": "eu-west-2",
		"vm_type": "t2.micro",
		"source_omi": "ami-65efcc11",
		"ssh_username": "outscale",
		"omi_name": "packer-test {{timestamp}}",
		"omi_virtualization_type": "hvm",
		"subregion_name": "eu-west-2a",
		"launch_block_device_mappings" : [
			{
			"volume_type" : "io1",
			"device_name" : "/dev/xvdf",
			"delete_on_vm_deletion" : false,
			"volume_size" : 10,
			"iops": 300,
			"no_device": 0
			}
		],
		"omi_root_device":{
			"source_device_name": "/dev/xvdf",
			"device_name": "/dev/sda1",
			"delete_on_vm_deletion": true,
			"volume_size": 10,
			"volume_type": "standard"
		}

	}]
}
`
