package ebssurrogate

import (
	"reflect"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const sourceDeviceName = "/dev/xvdf"
const rootDeviceName = "/dev/xvda"

func newStepRegisterAMI(amiDevices, launchDevices []*ec2.BlockDeviceMapping) *StepRegisterAMI {
	return &StepRegisterAMI{
		RootDevice: RootBlockDevice{
			SourceDeviceName:    sourceDeviceName,
			DeviceName:          rootDeviceName,
			DeleteOnTermination: true,
			VolumeType:          "ebs",
			VolumeSize:          10,
		},
		AMIDevices:    amiDevices,
		LaunchDevices: launchDevices,
	}
}

func sorted(devices []*ec2.BlockDeviceMapping) []*ec2.BlockDeviceMapping {
	sort.SliceStable(devices, func(i, j int) bool {
		return *devices[i].DeviceName < *devices[j].DeviceName
	})
	return devices
}

func TestStepRegisterAmi_combineDevices(t *testing.T) {
	cases := []struct {
		snapshotIds   map[string]string
		amiDevices    []*ec2.BlockDeviceMapping
		launchDevices []*ec2.BlockDeviceMapping
		allDevices    []*ec2.BlockDeviceMapping
	}{
		{
			snapshotIds:   map[string]string{},
			amiDevices:    []*ec2.BlockDeviceMapping{},
			launchDevices: []*ec2.BlockDeviceMapping{},
			allDevices:    []*ec2.BlockDeviceMapping{},
		},
		{
			snapshotIds: map[string]string{},
			amiDevices:  []*ec2.BlockDeviceMapping{},
			launchDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs:        &ec2.EbsBlockDevice{},
					DeviceName: aws.String(sourceDeviceName),
				},
			},
			allDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs:        &ec2.EbsBlockDevice{},
					DeviceName: aws.String(rootDeviceName),
				},
			},
		},
		{
			// Minimal single device
			snapshotIds: map[string]string{
				sourceDeviceName: "snap-0123456789abcdef1",
			},
			amiDevices: []*ec2.BlockDeviceMapping{},
			launchDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs:        &ec2.EbsBlockDevice{},
					DeviceName: aws.String(sourceDeviceName),
				},
			},
			allDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs: &ec2.EbsBlockDevice{
						SnapshotId: aws.String("snap-0123456789abcdef1"),
					},
					DeviceName: aws.String(rootDeviceName),
				},
			},
		},
		{
			// Single launch device with AMI device
			snapshotIds: map[string]string{
				sourceDeviceName: "snap-0123456789abcdef1",
			},
			amiDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs:        &ec2.EbsBlockDevice{},
					DeviceName: aws.String("/dev/xvdg"),
				},
			},
			launchDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs:        &ec2.EbsBlockDevice{},
					DeviceName: aws.String(sourceDeviceName),
				},
			},
			allDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs: &ec2.EbsBlockDevice{
						SnapshotId: aws.String("snap-0123456789abcdef1"),
					},
					DeviceName: aws.String(rootDeviceName),
				},
				{
					Ebs:        &ec2.EbsBlockDevice{},
					DeviceName: aws.String("/dev/xvdg"),
				},
			},
		},
		{
			// Multiple launch devices
			snapshotIds: map[string]string{
				sourceDeviceName: "snap-0123456789abcdef1",
				"/dev/xvdg":      "snap-0123456789abcdef2",
			},
			amiDevices: []*ec2.BlockDeviceMapping{},
			launchDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs:        &ec2.EbsBlockDevice{},
					DeviceName: aws.String(sourceDeviceName),
				},
				{
					Ebs:        &ec2.EbsBlockDevice{},
					DeviceName: aws.String("/dev/xvdg"),
				},
			},
			allDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs: &ec2.EbsBlockDevice{
						SnapshotId: aws.String("snap-0123456789abcdef1"),
					},
					DeviceName: aws.String(rootDeviceName),
				},
				{
					Ebs: &ec2.EbsBlockDevice{
						SnapshotId: aws.String("snap-0123456789abcdef2"),
					},
					DeviceName: aws.String("/dev/xvdg"),
				},
			},
		},
		{
			// Multiple launch devices with encryption
			snapshotIds: map[string]string{
				sourceDeviceName: "snap-0123456789abcdef1",
				"/dev/xvdg":      "snap-0123456789abcdef2",
			},
			amiDevices: []*ec2.BlockDeviceMapping{},
			launchDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs: &ec2.EbsBlockDevice{
						Encrypted: aws.Bool(true),
					},
					DeviceName: aws.String(sourceDeviceName),
				},
				{
					Ebs: &ec2.EbsBlockDevice{
						Encrypted: aws.Bool(true),
					},
					DeviceName: aws.String("/dev/xvdg"),
				},
			},
			allDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs: &ec2.EbsBlockDevice{
						SnapshotId: aws.String("snap-0123456789abcdef1"),
						// Encrypted: true stripped from snapshotted devices
					},
					DeviceName: aws.String(rootDeviceName),
				},
				{
					Ebs: &ec2.EbsBlockDevice{
						SnapshotId: aws.String("snap-0123456789abcdef2"),
					},
					DeviceName: aws.String("/dev/xvdg"),
				},
			},
		},
		{
			// Multiple launch devices and AMI devices with encryption
			snapshotIds: map[string]string{
				sourceDeviceName: "snap-0123456789abcdef1",
				"/dev/xvdg":      "snap-0123456789abcdef2",
			},
			amiDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs: &ec2.EbsBlockDevice{
						Encrypted: aws.Bool(true),
						KmsKeyId:  aws.String("keyId"),
					},
					// Source device name can be used in AMI devices
					// since launch device of same name gets renamed
					// to root device name
					DeviceName: aws.String(sourceDeviceName),
				},
			},
			launchDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs: &ec2.EbsBlockDevice{
						Encrypted: aws.Bool(true),
					},
					DeviceName: aws.String(sourceDeviceName),
				},
				{
					Ebs: &ec2.EbsBlockDevice{
						Encrypted: aws.Bool(true),
					},
					DeviceName: aws.String("/dev/xvdg"),
				},
			},
			allDevices: []*ec2.BlockDeviceMapping{
				{
					Ebs: &ec2.EbsBlockDevice{
						Encrypted: aws.Bool(true),
						KmsKeyId:  aws.String("keyId"),
					},
					DeviceName: aws.String(sourceDeviceName),
				},
				{
					Ebs: &ec2.EbsBlockDevice{
						SnapshotId: aws.String("snap-0123456789abcdef1"),
					},
					DeviceName: aws.String(rootDeviceName),
				},
				{
					Ebs: &ec2.EbsBlockDevice{
						SnapshotId: aws.String("snap-0123456789abcdef2"),
					},
					DeviceName: aws.String("/dev/xvdg"),
				},
			},
		},
	}
	for _, tc := range cases {
		stepRegisterAmi := newStepRegisterAMI(tc.amiDevices, tc.launchDevices)
		allDevices := stepRegisterAmi.combineDevices(tc.snapshotIds)
		if !reflect.DeepEqual(sorted(allDevices), sorted(tc.allDevices)) {
			t.Fatalf("Unexpected output from combineDevices")
		}
	}
}
