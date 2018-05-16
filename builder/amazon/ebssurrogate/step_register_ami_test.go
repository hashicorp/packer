package ebssurrogate

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetStringPointer() *string {
	tmp := "/dev/name"
	return &tmp
}

func GetTestDevice() *ec2.BlockDeviceMapping {
	TestDev := ec2.BlockDeviceMapping{
		DeviceName: GetStringPointer(),
	}
	return &TestDev
}

func TestStepRegisterAmi_DeduplicateRootVolume(t *testing.T) {
	TestRootDevice := RootBlockDevice{}
	TestRootDevice.SourceDeviceName = "/dev/name"

	blockDevices := []*ec2.BlockDeviceMapping{}
	blockDevicesExcludingRoot := DeduplicateRootVolume(blockDevices, TestRootDevice, "12342351")
	if len(blockDevicesExcludingRoot) != 1 {
		t.Fatalf("Unexpected length of block devices list")
	}

	TestBlockDevice := GetTestDevice()
	blockDevices = append(blockDevices, TestBlockDevice)
	blockDevicesExcludingRoot = DeduplicateRootVolume(blockDevices, TestRootDevice, "12342351")
	if len(blockDevicesExcludingRoot) != 1 {
		t.Fatalf("Unexpected length of block devices list")
	}
}
