package chroot

import (
	"testing"

	amazon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func testImage() ec2.Image {
	return ec2.Image{
		ImageId:      aws.String("ami-abcd1234"),
		Name:         aws.String("ami_test_name"),
		Architecture: aws.String("x86_64"),
		KernelId:     aws.String("aki-abcd1234"),
	}
}

func TestStepRegisterAmi_buildRegisterOpts_pv(t *testing.T) {
	config := Config{}
	config.AMIName = "test_ami_name"
	config.AMIDescription = "test_ami_description"
	config.AMIVirtType = "paravirtual"
	rootDeviceName := "foo"

	image := testImage()

	blockDevices := []*ec2.BlockDeviceMapping{}

	opts := buildRegisterOptsFromExistingImage(&config, &image, blockDevices, rootDeviceName)

	expected := config.AMIVirtType
	if *opts.VirtualizationType != expected {
		t.Fatalf("Unexpected VirtType value: expected %s got %s\n", expected, *opts.VirtualizationType)
	}

	expected = config.AMIName
	if *opts.Name != expected {
		t.Fatalf("Unexpected Name value: expected %s got %s\n", expected, *opts.Name)
	}

	expected = *image.KernelId
	if *opts.KernelId != expected {
		t.Fatalf("Unexpected KernelId value: expected %s got %s\n", expected, *opts.KernelId)
	}

	expected = rootDeviceName
	if *opts.RootDeviceName != expected {
		t.Fatalf("Unexpected RootDeviceName value: expected %s got %s\n", expected, *opts.RootDeviceName)
	}
}

func TestStepRegisterAmi_buildRegisterOpts_hvm(t *testing.T) {
	config := Config{}
	config.AMIName = "test_ami_name"
	config.AMIDescription = "test_ami_description"
	config.AMIVirtType = "hvm"
	rootDeviceName := "foo"

	image := testImage()

	blockDevices := []*ec2.BlockDeviceMapping{}

	opts := buildRegisterOptsFromExistingImage(&config, &image, blockDevices, rootDeviceName)

	expected := config.AMIVirtType
	if *opts.VirtualizationType != expected {
		t.Fatalf("Unexpected VirtType value: expected %s got %s\n", expected, *opts.VirtualizationType)
	}

	expected = config.AMIName
	if *opts.Name != expected {
		t.Fatalf("Unexpected Name value: expected %s got %s\n", expected, *opts.Name)
	}

	if opts.KernelId != nil {
		t.Fatalf("Unexpected KernelId value: expected nil got %s\n", *opts.KernelId)
	}

	expected = rootDeviceName
	if *opts.RootDeviceName != expected {
		t.Fatalf("Unexpected RootDeviceName value: expected %s got %s\n", expected, *opts.RootDeviceName)
	}
}

func TestStepRegisterAmi_buildRegisterOptsFromScratch(t *testing.T) {
	rootDeviceName := "/dev/sda"
	snapshotID := "foo"
	config := Config{
		FromScratch:  true,
		PackerConfig: common.PackerConfig{},
		AMIBlockDevices: amazon.AMIBlockDevices{
			AMIMappings: []amazon.BlockDevice{
				{
					DeviceName: rootDeviceName,
				},
			},
		},
		RootDeviceName: rootDeviceName,
	}
	registerOpts := buildBaseRegisterOpts(&config, nil, 10, snapshotID)

	if len(registerOpts.BlockDeviceMappings) != 1 {
		t.Fatal("Expected block device mapping of length 1")
	}

	if *registerOpts.BlockDeviceMappings[0].Ebs.SnapshotId != snapshotID {
		t.Fatalf("Snapshot ID of root disk not set to snapshot id %s", rootDeviceName)
	}
}

func TestStepRegisterAmi_buildRegisterOptFromExistingImage(t *testing.T) {
	rootDeviceName := "/dev/sda"
	snapshotID := "foo"

	config := Config{
		FromScratch:  false,
		PackerConfig: common.PackerConfig{},
	}
	sourceImage := ec2.Image{
		RootDeviceName: &rootDeviceName,
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: &rootDeviceName,
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(10),
				},
			},
			// Throw in an ephemeral device, it seems like all devices in the return struct in a source AMI have
			// a size, even if it's for ephemeral
			{
				DeviceName:  aws.String("/dev/sdb"),
				VirtualName: aws.String("ephemeral0"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(0),
				},
			},
		},
	}
	registerOpts := buildBaseRegisterOpts(&config, &sourceImage, 15, snapshotID)

	if len(registerOpts.BlockDeviceMappings) != 2 {
		t.Fatal("Expected block device mapping of length 2")
	}

	for _, dev := range registerOpts.BlockDeviceMappings {
		if dev.Ebs.SnapshotId != nil && *dev.Ebs.SnapshotId == snapshotID {
			// Even though root volume size is in config, it isn't used, instead we use the root volume size
			// that's derived when we build the step
			if *dev.Ebs.VolumeSize != 15 {
				t.Fatalf("Root volume size not 15 GB instead %d", *dev.Ebs.VolumeSize)
			}
			return
		}
	}
	t.Fatalf("Could not find device with snapshot ID %s", snapshotID)
}

func TestStepRegisterAmi_buildRegisterOptFromExistingImageWithBlockDeviceMappings(t *testing.T) {
	const (
		rootDeviceName = "/dev/xvda"
		oldRootDevice  = "/dev/sda1"
	)
	snapshotId := "foo"

	config := Config{
		FromScratch:  false,
		PackerConfig: common.PackerConfig{},
		AMIBlockDevices: amazon.AMIBlockDevices{
			AMIMappings: []amazon.BlockDevice{
				{
					DeviceName: rootDeviceName,
				},
			},
		},
		RootDeviceName: rootDeviceName,
	}

	// Intentionally try to use a different root devicename
	sourceImage := ec2.Image{
		RootDeviceName: aws.String(oldRootDevice),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String(oldRootDevice),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(10),
				},
			},
			// Throw in an ephemeral device, it seems like all devices in the return struct in a source AMI have
			// a size, even if it's for ephemeral
			{
				DeviceName:  aws.String("/dev/sdb"),
				VirtualName: aws.String("ephemeral0"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(0),
				},
			},
		},
	}
	registerOpts := buildBaseRegisterOpts(&config, &sourceImage, 15, snapshotId)

	if len(registerOpts.BlockDeviceMappings) != 1 {
		t.Fatal("Expected block device mapping of length 1")
	}

	if *registerOpts.BlockDeviceMappings[0].Ebs.SnapshotId != snapshotId {
		t.Fatalf("Snapshot ID of root disk set to '%s' expected '%s'", *registerOpts.BlockDeviceMappings[0].Ebs.SnapshotId, rootDeviceName)
	}

	if *registerOpts.RootDeviceName != rootDeviceName {
		t.Fatalf("Root device set to '%s' expected %s", *registerOpts.RootDeviceName, rootDeviceName)
	}

	if *registerOpts.BlockDeviceMappings[0].Ebs.VolumeSize != 15 {
		t.Fatalf("Size of root disk not set to 15 GB, instead %d", *registerOpts.BlockDeviceMappings[0].Ebs.VolumeSize)
	}
}
