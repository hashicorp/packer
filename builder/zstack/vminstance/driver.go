package vminstance

import (
	"github.com/hashicorp/packer/builder/zstack/zstacktype"
)

// Driver is the interface that has to be implemented to communicate
// with ZStack. The Driver interface exists mostly to allow a mock implementation
// to be used to test the steps.
type Driver interface {
	QueryZone(uuid string) (*zstacktype.Zone, error)
	QueryImage(uuid string) (*zstacktype.Image, error)
	QueryVm(uuid string) (*zstacktype.VmInstance, error)
	QueryBackupStorage(uuid string) (*zstacktype.BackupStorage, error)
	QueryL3Network(uuid string) (*zstacktype.L3Network, error)
	QueryVolume(uuid string) (*zstacktype.DataVolume, error)

	CreateVmInstance(zstacktype.CreateVm) (*zstacktype.VmInstance, error)
	StopVmInstance(uuid string) error
	DeleteVmInstance(uuid string) error
	// WaitForInstance waits for an instance to reach the given state.
	WaitForInstance(state, uuid, instanceType string) <-chan error

	CreateImage(zstacktype.CreateImage) (*zstacktype.Image, error)
	CreateDataVolumeImage(zstacktype.CreateVolumeImage) (*zstacktype.Image, error)
	ExportImage(zstacktype.Image) (string, error)

	CreateDataVolumeFromImage(zstacktype.CreateDataVolumeFromImage) (*zstacktype.DataVolume, error)
	CreateDataVolumeFromSize(zstacktype.CreateDataVolume) (*zstacktype.DataVolume, error)
	AttachDataVolume(uuid, vmUuid string) (string, error)
	DetachDataVolume(uuid string) error
	DeleteDataVolume(uuid string) error

	GetZStackVersion() (string, error)
}
