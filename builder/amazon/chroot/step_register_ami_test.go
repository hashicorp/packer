package chroot

import (
	"testing"

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

	image := testImage()

	blockDevices := []*ec2.BlockDeviceMapping{}

	opts := buildRegisterOpts(&config, &image, blockDevices)

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

}

func TestStepRegisterAmi_buildRegisterOpts_hvm(t *testing.T) {
	config := Config{}
	config.AMIName = "test_ami_name"
	config.AMIDescription = "test_ami_description"
	config.AMIVirtType = "hvm"

	image := testImage()

	blockDevices := []*ec2.BlockDeviceMapping{}

	opts := buildRegisterOpts(&config, &image, blockDevices)

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
}
