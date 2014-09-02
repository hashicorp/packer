package chroot

import (
	"github.com/mitchellh/goamz/ec2"
	"testing"
)

func testImage() ec2.Image {
	return ec2.Image{
		Id:           "ami-abcd1234",
		Name:         "ami_test_name",
		Architecture: "x86_64",
		KernelId:     "aki-abcd1234",
	}
}

func TestStepRegisterAmi_buildRegisterOpts_pv(t *testing.T) {
	config := Config{}
	config.AMIName = "test_ami_name"
	config.AMIDescription = "test_ami_description"
	config.AMIVirtType = "paravirtual"

	image := testImage()

	blockDevices := []ec2.BlockDeviceMapping{}

	opts := buildRegisterOpts(&config, &image, blockDevices)

	expected := config.AMIVirtType
	if opts.VirtType != expected {
		t.Fatalf("Unexpected VirtType value: expected %s got %s\n", expected, opts.VirtType)
	}

	expected = config.AMIName
	if opts.Name != expected {
		t.Fatalf("Unexpected Name value: expected %s got %s\n", expected, opts.Name)
	}

	expected = image.KernelId
	if opts.KernelId != expected {
		t.Fatalf("Unexpected KernelId value: expected %s got %s\n", expected, opts.KernelId)
	}

}

func TestStepRegisterAmi_buildRegisterOpts_hvm(t *testing.T) {
	config := Config{}
	config.AMIName = "test_ami_name"
	config.AMIDescription = "test_ami_description"
	config.AMIVirtType = "hvm"

	image := testImage()

	blockDevices := []ec2.BlockDeviceMapping{}

	opts := buildRegisterOpts(&config, &image, blockDevices)

	expected := config.AMIVirtType
	if opts.VirtType != expected {
		t.Fatalf("Unexpected VirtType value: expected %s got %s\n", expected, opts.VirtType)
	}

	expected = config.AMIName
	if opts.Name != expected {
		t.Fatalf("Unexpected Name value: expected %s got %s\n", expected, opts.Name)
	}

	expected = ""
	if opts.KernelId != expected {
		t.Fatalf("Unexpected KernelId value: expected %s got %s\n", expected, opts.KernelId)
	}

}
