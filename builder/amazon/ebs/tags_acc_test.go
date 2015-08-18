package ebs

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/packer/builder/amazon/common"
	builderT "github.com/mitchellh/packer/helper/builder/testing"
	"github.com/mitchellh/packer/packer"
)

func TestBuilderTagsAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderTagsAccBasic,
		Check:    checkTags(),
	})
}

func checkTags() builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		if len(artifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		tags := make(map[string]string)
		tags["OS_Version"] = "Ubuntu"
		tags["Release"] = "Latest"

		// Get the actual *Artifact pointer so we can access the AMIs directly
		artifactRaw := artifacts[0]
		artifact, ok := artifactRaw.(*common.Artifact)
		if !ok {
			return fmt.Errorf("unknown artifact: %#v", artifactRaw)
		}

		// describe the image, get block devices with a snapshot
		ec2conn, _ := testEC2Conn()
		imageResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(artifact.Amis["us-east-1"])},
		})

		if err != nil {
			return fmt.Errorf("Error retrieving details for AMI Artifcat (%#v) in Tags Test: %s", artifact, err)
		}

		if len(imageResp.Images) == 0 {
			return fmt.Errorf("No images found for AMI Artifcat (%#v) in Tags Test: %s", artifact, err)
		}

		image := imageResp.Images[0]

		// Check only those with a Snapshot ID, i.e. not Ephemeral
		var snapshots []*string
		for _, device := range image.BlockDeviceMappings {
			if device.Ebs != nil && device.Ebs.SnapshotId != nil {
				snapshots = append(snapshots, device.Ebs.SnapshotId)
			}
		}

		// grab matching snapshot info
		resp, err := ec2conn.DescribeSnapshots(&ec2.DescribeSnapshotsInput{
			SnapshotIds: snapshots,
		})

		if err != nil {
			return fmt.Errorf("Error retreiving Snapshots for AMI Artifcat (%#v) in Tags Test: %s", artifact, err)
		}

		if len(resp.Snapshots) == 0 {
			return fmt.Errorf("No Snapshots found for AMI Artifcat (%#v) in Tags Test", artifact)
		}

		// grab the snapshots, check the tags
		for _, s := range resp.Snapshots {
			expected := len(tags)
			for _, t := range s.Tags {
				for key, value := range tags {
					if key == *t.Key && value == *t.Value {
						expected--
					}
				}
			}

			if expected > 0 {
				return fmt.Errorf("Not all tags found")
			}
		}

		return nil
	}
}

const testBuilderTagsAccBasic = `
{
  "builders": [
    {
      "type": "test",
      "region": "us-east-1",
      "source_ami": "ami-9eaa1cf6",
      "instance_type": "t2.micro",
      "ssh_username": "ubuntu",
      "ami_name": "packer-tags-testing-{{timestamp}}",
      "tags": {
        "OS_Version": "Ubuntu",
        "Release": "Latest"
      }
    }
  ]
}
`
