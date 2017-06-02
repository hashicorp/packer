package ecs

import (
	"os"
	"testing"

	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	"github.com/hashicorp/packer/packer"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

//func TestBuilderAcc_windows(t *testing.T) {
//	builderT.Test(t, builderT.TestCase{
//		PreCheck: func() {
//			testAccPreCheck(t)
//		},
//		Builder:  &Builder{},
//		Template: testBuilderAccWindows,
//	})
//}

//func TestBuilderAcc_regionCopy(t *testing.T) {
//	builderT.Test(t, builderT.TestCase{
//		PreCheck: func() {
//			testAccPreCheck(t)
//		},
//		Builder:  &Builder{},
//		Template: testBuilderAccRegionCopy,
//		Check:    checkRegionCopy([]string{"cn-hangzhou", "cn-shenzhen"}),
//	})
//}

func TestBuilderAcc_forceDelete(t *testing.T) {
	// Build the same alicloud image twice, with ecs_image_force_delete on the second run
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Builder:              &Builder{},
		Template:             buildForceDeregisterConfig("false", "delete"),
		SkipArtifactTeardown: true,
	})

	builderT.Test(t, builderT.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Builder:  &Builder{},
		Template: buildForceDeregisterConfig("true", "delete"),
	})
}

func TestBuilderAcc_ECSImageSharing(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Builder:  &Builder{},
		Template: testBuilderAccSharing,
		Check:    checkECSImageSharing("1309208528360047"),
	})
}

func TestBuilderAcc_forceDeleteSnapshot(t *testing.T) {
	destImageName := "delete"

	// Build the same alicloud image name twice, with force_delete_snapshot on the second run
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Builder:              &Builder{},
		Template:             buildForceDeleteSnapshotConfig("false", destImageName),
		SkipArtifactTeardown: true,
	})

	// Get image data by image image name
	client, _ := testAliyunClient()
	images, _, _ := client.DescribeImages(&ecs.DescribeImagesArgs{
		ImageName: "packer-test-" + destImageName,
		RegionId:  common.Region("cn-beijing")})

	image := images[0]

	// Get snapshot ids for image
	snapshotIds := []string{}
	for _, device := range image.DiskDeviceMappings.DiskDeviceMapping {
		if device.Device != "" && device.SnapshotId != "" {
			snapshotIds = append(snapshotIds, device.SnapshotId)
		}
	}

	builderT.Test(t, builderT.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Builder:  &Builder{},
		Template: buildForceDeleteSnapshotConfig("true", destImageName),
		Check:    checkSnapshotsDeleted(snapshotIds),
	})
}

func checkSnapshotsDeleted(snapshotIds []string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		// Verify the snapshots are gone
		client, _ := testAliyunClient()
		snapshotResp, _, err := client.DescribeSnapshots(
			&ecs.DescribeSnapshotsArgs{RegionId: common.Region("cn-beijing"), SnapshotIds: snapshotIds},
		)
		if err != nil {
			return fmt.Errorf("Query snapshot failed %v", err)
		}
		if len(snapshotResp) > 0 {
			return fmt.Errorf("Snapshots weren't successfully deleted by " +
				"`ecs_image_force_delete_snapshots`")
		}
		return nil
	}
}

func checkECSImageSharing(uid string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		if len(artifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		// Get the actual *Artifact pointer so we can access the AMIs directly
		artifactRaw := artifacts[0]
		artifact, ok := artifactRaw.(*Artifact)
		if !ok {
			return fmt.Errorf("unknown artifact: %#v", artifactRaw)
		}

		// describe the image, get block devices with a snapshot
		client, _ := testAliyunClient()
		imageSharePermissionResponse, err := client.DescribeImageSharePermission(
			&ecs.ModifyImageSharePermissionArgs{
				RegionId: "cn-beijing",
				ImageId:  artifact.AlicloudImages["cn-beijing"],
			})

		if err != nil {
			return fmt.Errorf("Error retrieving Image Attributes for ECS Image Artifact (%#v) "+
				"in ECS Image Sharing Test: %s", artifact, err)
		}

		if len(imageSharePermissionResponse.Accounts.Account) != 1 &&
			imageSharePermissionResponse.Accounts.Account[0].AliyunId != uid {
			return fmt.Errorf("share account is incorrect %d",
				len(imageSharePermissionResponse.Accounts.Account))
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
		artifact, ok := artifactRaw.(*Artifact)
		if !ok {
			return fmt.Errorf("unknown artifact: %#v", artifactRaw)
		}

		// Verify that we copied to only the regions given
		regionSet := make(map[string]struct{})
		for _, r := range regions {
			regionSet[r] = struct{}{}
		}
		for r := range artifact.AlicloudImages {
			if r == "cn-beijing" {
				delete(regionSet, r)
				continue
			}
			if _, ok := regionSet[r]; !ok {
				return fmt.Errorf("unknown region: %s", r)
			}

			delete(regionSet, r)
		}
		if len(regionSet) > 0 {
			return fmt.Errorf("didn't copy to: %#v", regionSet)
		}
		client, _ := testAliyunClient()
		for key, value := range artifact.AlicloudImages {
			client.WaitForImageReady(common.Region(key), value, 1800)
		}
		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ALICLOUD_ACCESS_KEY"); v == "" {
		t.Fatal("ALICLOUD_ACCESS_KEY must be set for acceptance tests")
	}

	if v := os.Getenv("ALICLOUD_SECRET_KEY"); v == "" {
		t.Fatal("ALICLOUD_SECRET_KEY must be set for acceptance tests")
	}
}

func testAliyunClient() (*ecs.Client, error) {
	access := &AlicloudAccessConfig{AlicloudRegion: "cn-beijing"}
	err := access.Config()
	if err != nil {
		return nil, err
	}
	client, err := access.Client()
	if err != nil {
		return nil, err
	}

	return client, nil
}

const testBuilderAccBasic = `
{	"builders": [{
		"type": "test",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_16_0402_64_40G_base_20170222.vhd",
		"ssh_username": "ubuntu",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test_{{timestamp}}"
	}]
}`

const testBuilderAccRegionCopy = `
{
	"builders": [{
		"type": "test",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_16_0402_64_40G_base_20170222.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test_{{timestamp}}",
		"image_copy_regions": ["cn-hangzhou", "cn-shenzhen"]
	}]
}
`

const testBuilderAccForceDelete = `
{
	"builders": [{
		"type": "test",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_16_0402_64_40G_base_20170222.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_force_delete": "%s",
		"image_name": "packer-test_%s"
	}]
}
`

const testBuilderAccForceDeleteSnapshot = `
{
	"builders": [{
		"type": "test",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_16_0402_64_40G_base_20170222.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_force_delete_snapshots": "%s",
		"image_force_delete": "%s",
		"image_name": "packer-test-%s"
	}]
}
`

// share with catsby
const testBuilderAccSharing = `
{
	"builders": [{
		"type": "test",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"ubuntu_16_0402_64_40G_base_20170222.vhd",
		"io_optimized":"true",
		"ssh_username":"root",
		"image_name": "packer-test_{{timestamp}}",
		"image_share_account":["1309208528360047"]
	}]
}
`

func buildForceDeregisterConfig(val, name string) string {
	return fmt.Sprintf(testBuilderAccForceDelete, val, name)
}

func buildForceDeleteSnapshotConfig(val, name string) string {
	return fmt.Sprintf(testBuilderAccForceDeleteSnapshot, val, val, name)
}

const testBuilderAccWindows = `
{	"builders": [{
		"type": "test",
		"region": "cn-beijing",
		"instance_type": "ecs.n1.tiny",
		"source_image":"win2008_64_ent_r2_zh-cn_40G_alibase_20170301.vhd",
		"io_optimized":"true",
		"image_force_delete":"true",
		"communicator": "winrm",
		"winrm_port": 5985,
		"winrm_username": "Administrator",
		"winrm_password": "Test1234",
		"image_name": "packer-test_{{timestamp}}"
	}]
}`
