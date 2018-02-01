package ebssurrogate

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/template/interpolate"
)

type RootBlockDevice struct {
	SourceDeviceName    string `mapstructure:"source_device_name"`
	DeviceName          string `mapstructure:"device_name"`
	DeleteOnTermination bool   `mapstructure:"delete_on_termination"`
	IOPS                int64  `mapstructure:"iops"`
	VolumeType          string `mapstructure:"volume_type"`
	VolumeSize          int64  `mapstructure:"volume_size"`
}

func (c *RootBlockDevice) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.SourceDeviceName == "" {
		errs = append(errs, errors.New("source_device_name for the root_device must be specified"))
	}

	if c.DeviceName == "" {
		errs = append(errs, errors.New("device_name for the root_device must be specified"))
	}

	if c.VolumeType == "gp2" && c.IOPS != 0 {
		errs = append(errs, errors.New("iops may not be specified for a gp2 volume"))
	}

	if c.IOPS < 0 {
		errs = append(errs, errors.New("iops must be greater than 0"))
	}

	if c.VolumeSize < 0 {
		errs = append(errs, errors.New("volume_size must be greater than 0"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (d *RootBlockDevice) createBlockDeviceMapping(snapshotId string) *ec2.BlockDeviceMapping {
	rootBlockDevice := &ec2.EbsBlockDevice{
		SnapshotId:          aws.String(snapshotId),
		VolumeType:          aws.String(d.VolumeType),
		VolumeSize:          aws.Int64(d.VolumeSize),
		DeleteOnTermination: aws.Bool(d.DeleteOnTermination),
	}

	if d.IOPS != 0 {
		rootBlockDevice.Iops = aws.Int64(d.IOPS)
	}

	return &ec2.BlockDeviceMapping{
		DeviceName: aws.String(d.DeviceName),
		Ebs:        rootBlockDevice,
	}
}
