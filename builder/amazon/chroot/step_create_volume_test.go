package chroot

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func buildTestRootDevice() *ec2.BlockDeviceMapping {
	return &ec2.BlockDeviceMapping{
		Ebs: &ec2.EbsBlockDevice{
			VolumeSize: aws.Int64(10),
			SnapshotId: aws.String("snap-1234"),
			VolumeType: aws.String("gp2"),
		},
	}
}

func TestCreateVolume_Default(t *testing.T) {
	stepCreateVolume := new(StepCreateVolume)
	_, err := stepCreateVolume.buildCreateVolumeInput("test-az", buildTestRootDevice())
	assert.NoError(t, err)
}

func TestCreateVolume_Shrink(t *testing.T) {
	stepCreateVolume := StepCreateVolume{RootVolumeSize: 1}
	testRootDevice := buildTestRootDevice()
	ret, err := stepCreateVolume.buildCreateVolumeInput("test-az", testRootDevice)
	assert.NoError(t, err)
	// Ensure that the new value is equal to the size of the old root device
	assert.Equal(t, *ret.Size, *testRootDevice.Ebs.VolumeSize)
}

func TestCreateVolume_Expand(t *testing.T) {
	stepCreateVolume := StepCreateVolume{RootVolumeSize: 25}
	testRootDevice := buildTestRootDevice()
	ret, err := stepCreateVolume.buildCreateVolumeInput("test-az", testRootDevice)
	assert.NoError(t, err)
	// Ensure that the new value is equal to the size of the value passed in
	assert.Equal(t, *ret.Size, stepCreateVolume.RootVolumeSize)
}

func TestCreateVolume_io1_to_io1(t *testing.T) {
	stepCreateVolume := StepCreateVolume{RootVolumeType: "io1"}
	testRootDevice := buildTestRootDevice()
	testRootDevice.Ebs.VolumeType = aws.String("io1")
	testRootDevice.Ebs.Iops = aws.Int64(1000)
	ret, err := stepCreateVolume.buildCreateVolumeInput("test-az", testRootDevice)
	assert.NoError(t, err)
	assert.Equal(t, *ret.VolumeType, stepCreateVolume.RootVolumeType)
	assert.Equal(t, *ret.Iops, *testRootDevice.Ebs.Iops)
}

func TestCreateVolume_io1_to_gp2(t *testing.T) {
	stepCreateVolume := StepCreateVolume{RootVolumeType: "gp2"}
	testRootDevice := buildTestRootDevice()
	testRootDevice.Ebs.VolumeType = aws.String("io1")
	testRootDevice.Ebs.Iops = aws.Int64(1000)

	ret, err := stepCreateVolume.buildCreateVolumeInput("test-az", testRootDevice)
	assert.NoError(t, err)
	assert.Equal(t, *ret.VolumeType, stepCreateVolume.RootVolumeType)
	assert.Nil(t, ret.Iops)
}

func TestCreateVolume_gp2_to_io1(t *testing.T) {
	stepCreateVolume := StepCreateVolume{RootVolumeType: "io1"}
	testRootDevice := buildTestRootDevice()

	_, err := stepCreateVolume.buildCreateVolumeInput("test-az", testRootDevice)
	assert.Error(t, err)
}
