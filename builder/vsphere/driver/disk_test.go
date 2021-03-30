package driver

import (
	"testing"

	"github.com/vmware/govmomi/object"
)

func TestAddStorageDevices(t *testing.T) {
	config := &StorageConfig{
		DiskControllerType: []string{"pvscsi"},
		Storage: []Disk{
			{
				DiskSize:            3072,
				DiskThinProvisioned: true,
				ControllerIndex:     0,
			},
			{
				DiskSize:            20480,
				DiskThinProvisioned: true,
				ControllerIndex:     0,
			},
		},
	}

	noExistingDevices := object.VirtualDeviceList{}
	storageConfigSpec, err := config.AddStorageDevices(noExistingDevices)
	if err != nil {
		t.Fatalf("unexpected erro: %q", err.Error())
	}
	if len(storageConfigSpec) != 3 {
		t.Fatalf("Expecting VirtualDeviceList to have 3 storage devices but had %d", len(storageConfigSpec))
	}

	existingDevices := object.VirtualDeviceList{}
	device, err := existingDevices.CreateNVMEController()
	existingDevices = append(existingDevices, device)

	storageConfigSpec, err = config.AddStorageDevices(existingDevices)
	if err != nil {
		t.Fatalf("unexpected erro: %q", err.Error())
	}
	if len(storageConfigSpec) != 3 {
		t.Fatalf("Expecting VirtualDeviceList to have 3 storage devices but had %d", len(storageConfigSpec))
	}
}
