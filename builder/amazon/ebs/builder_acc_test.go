/*
Deregister the test image with
aws ec2 deregister-image --image-id $(aws ec2 describe-images --output text --filters "Name=name,Values=packer-test-packer-test-dereg" --query 'Images[*].{ID:ImageId}')
*/
package ebs

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/builder/amazon/common"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
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

func TestBuilderAcc_forceDeleteSnapshot(t *testing.T) {
	amiName := "packer-test-dereg"

	// Build the same AMI name twice, with force_delete_snapshot on the second run
	builderT.Test(t, builderT.TestCase{
		PreCheck:             func() { testAccPreCheck(t) },
		Builder:              &Builder{},
		Template:             buildForceDeleteSnapshotConfig("false", amiName),
		SkipArtifactTeardown: true,
	})

	// Get image data by AMI name
	ec2conn, _ := testEC2Conn()
	describeInput := &ec2.DescribeImagesInput{Filters: []*ec2.Filter{
		{
			Name:   aws.String("name"),
			Values: []*string{aws.String(amiName)},
		},
	}}
	ec2conn.WaitUntilImageExists(describeInput)
	imageResp, _ := ec2conn.DescribeImages(describeInput)
	image := imageResp.Images[0]

	// Get snapshot ids for image
	snapshotIds := []*string{}
	for _, device := range image.BlockDeviceMappings {
		if device.Ebs != nil && device.Ebs.SnapshotId != nil {
			snapshotIds = append(snapshotIds, device.Ebs.SnapshotId)
		}
	}

	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: buildForceDeleteSnapshotConfig("true", amiName),
		Check:    checkSnapshotsDeleted(snapshotIds),
	})
}

func checkSnapshotsDeleted(snapshotIds []*string) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		// Verify the snapshots are gone
		ec2conn, _ := testEC2Conn()
		snapshotResp, _ := ec2conn.DescribeSnapshots(
			&ec2.DescribeSnapshotsInput{SnapshotIds: snapshotIds},
		)

		if len(snapshotResp.Snapshots) > 0 {
			return fmt.Errorf("Snapshots weren't successfully deleted by `force_delete_snapshot`")
		}
		return nil
	}
}

func TestBuilderAcc_amiSharing(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccSharingPreCheck(t) },
		Builder:  &Builder{},
		Template: buildSharingConfig(os.Getenv("TESTACC_AWS_ACCOUNT_ID")),
		Check:    checkAMISharing(2, os.Getenv("TESTACC_AWS_ACCOUNT_ID"), "all"),
	})
}

func TestBuilderAcc_encryptedBoot(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccEncrypted,
		Check:    checkBootEncrypted(),
	})
}

func checkAMISharing(count int, uid, group string) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
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
	return func(artifacts []packersdk.Artifact) error {
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
		for r := range artifact.Amis {
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

func checkBootEncrypted() builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {

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
			return fmt.Errorf("Error retrieving Image Attributes for AMI (%s) in AMI Encrypted Boot Test: %s", artifact, err)
		}

		image := imageResp.Images[0] // Only requested a single AMI ID

		rootDeviceName := image.RootDeviceName

		for _, bd := range image.BlockDeviceMappings {
			if *bd.DeviceName == *rootDeviceName {
				if *bd.Ebs.Encrypted != true {
					return fmt.Errorf("volume not encrypted: %s", *bd.Ebs.SnapshotId)
				}
			}
		}

		return nil
	}
}

func TestBuilderAcc_SessionManagerInterface(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccSessionManagerInterface,
	})
}

func testAccPreCheck(t *testing.T) {
}

func testAccSharingPreCheck(t *testing.T) {
	if v := os.Getenv("TESTACC_AWS_ACCOUNT_ID"); v == "" {
		t.Fatal(fmt.Sprintf("TESTACC_AWS_ACCOUNT_ID must be set for acceptance tests"))
	}
}

func testEC2Conn() (*ec2.EC2, error) {
	access := &common.AccessConfig{RawRegion: "us-east-1"}
	session, err := access.Session()
	if err != nil {
		return nil, err
	}

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
		"ami_name": "%s"
	}]
}
`

const testBuilderAccForceDeleteSnapshot = `
{
	"builders": [{
		"type": "test",
		"region": "us-east-1",
		"instance_type": "m3.medium",
		"source_ami": "ami-76b2a71e",
		"ssh_username": "ubuntu",
		"force_deregister": "%s",
		"force_delete_snapshot": "%s",
		"ami_name": "%s"
	}]
}
`

const testBuilderAccSharing = `
{
	"builders": [{
		"type": "test",
		"region": "us-east-1",
		"instance_type": "m3.medium",
		"source_ami": "ami-76b2a71e",
		"ssh_username": "ubuntu",
		"ami_name": "packer-test {{timestamp}}",
		"ami_users":["%s"],
		"ami_groups":["all"]
	}]
}
`

const testBuilderAccEncrypted = `
{
	"builders": [{
		"type": "test",
		"region": "us-east-1",
		"instance_type": "m3.medium",
		"source_ami":"ami-c15bebaa",
		"ssh_username": "ubuntu",
		"ami_name": "packer-enc-test {{timestamp}}",
		"encrypt_boot": true
	}]
}
`

const testBuilderAccSessionManagerInterface = `
{
	"builders": [{
		"type": "test",
		"region": "us-east-1",
		"instance_type": "m3.medium",
		"source_ami_filter": {
				"filters": {
						"virtualization-type": "hvm",
						"name": "ubuntu/images/*ubuntu-xenial-16.04-amd64-server-*",
						"root-device-type": "ebs"
				},
				"owners": [
						"099720109477"
				],
				"most_recent": true
		},
		"ssh_username": "ubuntu",
		"ssh_interface": "session_manager",
		"iam_instance_profile": "SSMInstanceProfile",
		"ami_name": "packer-ssm-test-{{timestamp}}"
	}]
}
`

func buildForceDeregisterConfig(val, name string) string {
	return fmt.Sprintf(testBuilderAccForceDeregister, val, name)
}

func buildForceDeleteSnapshotConfig(val, name string) string {
	return fmt.Sprintf(testBuilderAccForceDeleteSnapshot, val, val, name)
}

func buildSharingConfig(val string) string {
	return fmt.Sprintf(testBuilderAccSharing, val)
}
