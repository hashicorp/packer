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
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

func testAccPreCheck(t *testing.T) {
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"region": "eu-west-2",
		"vm_type": "c4.large",
		"source_omi": "ami-46260446",
		"ssh_username": "ubuntu",
		"omi_name": "packer-test {{timestamp}}",
		"omi_virtualization_type": "hvm",
		"subregion_name": "eu-west-2a",
		"launch_block_device_mappings" : [
			{
			"volume_type" : "io1",
			"device_name" : "/dev/xvdf",
			"delete_on_vm_deletion" : false,
			"volume_size" : 10,
			"iops": 300
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
