package ebs

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/packer/builder/amazon/common"
	builderT "github.com/mitchellh/packer/helper/builder/testing"
	"github.com/mitchellh/packer/packer"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

func TestBuilderAcc_regionCopy(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccRegionCopy,
		Check:    checkRegionCopy([]string{"us-east-1", "us-west-2"}),
	})
}

func TestBuilderAcc_forceDeregister(t *testing.T) {
	// Build the same AMI name twice, with force_deregister on the second run
	builderT.Test(t, builderT.TestCase{
		PreCheck:             func() { testAccPreCheck(t) },
		Builder:              &Builder{},
		Template:             buildForceDeregisterConfig("false", "dereg"),
		SkipArtifactTeardown: true,
	})

	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: buildForceDeregisterConfig("true", "dereg"),
	})
}

func TestBuilderAcc_amiSharing(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccSharing,
		Check:    checkAMISharing(2, "932021504756", "all"),
	})
}

func checkAMISharing(count int, uid, group string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		if len(artifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		// Get the actual *Artifact pointer so we can access the AMIs directly
		artifactRaw := artifacts[0]
		artifact, ok := artifactRaw.(*common.Artifact)
		if !ok {
			return fmt.Errorf("unknown artifact: %#v", artifactRaw)
		}

		// describe the image, get block devices with a snapshot
		ec2conn, _ := testEC2Conn()
		imageResp, err := ec2conn.DescribeImageAttribute(&ec2.DescribeImageAttributeInput{
			Attribute: aws.String("launchPermission"),
			ImageId:   aws.String(artifact.Amis["us-east-1"]),
		})

		if err != nil {
			return fmt.Errorf("Error retrieving Image Attributes for AMI Artifact (%#v) in AMI Sharing Test: %s", artifact, err)
		}

		// Launch Permissions are in addition to the userid that created it, so if
		// you add 3 additional ami_users, you expect 2 Launch Permissions here
		if len(imageResp.LaunchPermissions) != count {
			return fmt.Errorf("Error in Image Attributes, expected (%d) Launch Permissions, got (%d)", count, len(imageResp.LaunchPermissions))
		}

		userFound := false
		for _, lp := range imageResp.LaunchPermissions {
			if lp.UserId != nil && uid == *lp.UserId {
				userFound = true
			}
		}

		if !userFound {
			return fmt.Errorf("Error in Image Attributes, expected User ID (%s) to have Launch Permissions, but was not found", uid)
		}

		groupFound := false
		for _, lp := range imageResp.LaunchPermissions {
			if lp.Group != nil && group == *lp.Group {
				groupFound = true
			}
		}

		if !groupFound {
			return fmt.Errorf("Error in Image Attributes, expected Group ID (%s) to have Launch Permissions, but was not found", group)
		}

		return nil
	}
}

func checkRegionCopy(regions []string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		if len(artifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		// Get the actual *Artifact pointer so we can access the AMIs directly
		artifactRaw := artifacts[0]
		artifact, ok := artifactRaw.(*common.Artifact)
		if !ok {
			return fmt.Errorf("unknown artifact: %#v", artifactRaw)
		}

		// Verify that we copied to only the regions given
		regionSet := make(map[string]struct{})
		for _, r := range regions {
			regionSet[r] = struct{}{}
		}
		for r, _ := range artifact.Amis {
			if _, ok := regionSet[r]; !ok {
				return fmt.Errorf("unknown region: %s", r)
			}

			delete(regionSet, r)
		}
		if len(regionSet) > 0 {
			return fmt.Errorf("didn't copy to: %#v", regionSet)
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("AWS_ACCESS_KEY_ID"); v == "" {
		t.Fatal("AWS_ACCESS_KEY_ID must be set for acceptance tests")
	}

	if v := os.Getenv("AWS_SECRET_ACCESS_KEY"); v == "" {
		t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
	}
}

func testEC2Conn() (*ec2.EC2, error) {
	access := &common.AccessConfig{RawRegion: "us-east-1"}
	config, err := access.Config()
	if err != nil {
		return nil, err
	}

	session := session.New(config)
	return ec2.New(session), nil
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"region": "us-east-1",
		"instance_type": "m3.medium",
		"source_ami": "ami-76b2a71e",
		"ssh_username": "ubuntu",
		"ami_name": "packer-test {{timestamp}}"
	}]
}
`

const testBuilderAccRegionCopy = `
{
	"builders": [{
		"type": "test",
		"region": "us-east-1",
		"instance_type": "m3.medium",
		"source_ami": "ami-76b2a71e",
		"ssh_username": "ubuntu",
		"ami_name": "packer-test {{timestamp}}",
		"ami_regions": ["us-east-1", "us-west-2"]
	}]
}
`

const testBuilderAccForceDeregister = `
{
	"builders": [{
		"type": "test",
		"region": "us-east-1",
		"instance_type": "m3.medium",
		"source_ami": "ami-76b2a71e",
		"ssh_username": "ubuntu",
		"force_deregister": "%s",
		"ami_name": "packer-test-%s"
	}]
}
`

// share with catsby
const testBuilderAccSharing = `
{
	"builders": [{
		"type": "test",
		"region": "us-east-1",
		"instance_type": "m3.medium",
		"source_ami": "ami-76b2a71e",
		"ssh_username": "ubuntu",
		"ami_name": "packer-test {{timestamp}}",
		"ami_users":["932021504756"],
		"ami_groups":["all"]
	}]
}
`

func buildForceDeregisterConfig(name, flag string) string {
	return fmt.Sprintf(testBuilderAccForceDeregister, name, flag)
}
