package driver

import (
	"fmt"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vim25/types"
)

type DriverMock struct {
	FindDatastoreCalled bool
	DatastoreMock       *DatastoreMock
	FindDatastoreName   string
	FindDatastoreHost   string
	FindDatastoreErr    error

	PreCleanShouldFail bool
	PreCleanVMCalled   bool
	PreCleanForce      bool
	PreCleanVMPath     string

	CreateVMShouldFail bool
	CreateVMCalled     bool
	CreateConfig       *CreateConfig
	VM                 VirtualMachine
}

func NewDriverMock() *DriverMock {
	return new(DriverMock)
}

func (d *DriverMock) FindDatastore(name string, host string) (Datastore, error) {
	d.FindDatastoreCalled = true
	if d.DatastoreMock == nil {
		d.DatastoreMock = new(DatastoreMock)
	}
	d.FindDatastoreName = name
	d.FindDatastoreHost = host
	return d.DatastoreMock, d.FindDatastoreErr
}

func (d *DriverMock) NewVM(ref *types.ManagedObjectReference) VirtualMachine {
	return nil
}

func (d *DriverMock) FindVM(name string) (VirtualMachine, error) {
	return nil, nil
}

func (d *DriverMock) FindCluster(name string) (*Cluster, error) {
	return nil, nil
}

func (d *DriverMock) PreCleanVM(ui packersdk.Ui, vmPath string, force bool) error {
	d.PreCleanVMCalled = true
	if d.PreCleanShouldFail {
		return fmt.Errorf("pre clean failed")
	}
	d.PreCleanForce = true
	d.PreCleanVMPath = vmPath
	return nil
}

func (d *DriverMock) CreateVM(config *CreateConfig) (VirtualMachine, error) {
	d.CreateVMCalled = true
	if d.CreateVMShouldFail {
		return nil, fmt.Errorf("create vm failed")
	}
	d.CreateConfig = config
	d.VM = new(VirtualMachineDriver)
	return d.VM, nil
}

func (d *DriverMock) NewDatastore(ref *types.ManagedObjectReference) Datastore { return nil }

func (d *DriverMock) GetDatastoreName(id string) (string, error) { return "", nil }

func (d *DriverMock) GetDatastoreFilePath(datastoreID, dir, filename string) (string, error) {
	return "", nil
}

func (d *DriverMock) NewFolder(ref *types.ManagedObjectReference) *Folder { return nil }

func (d *DriverMock) FindFolder(name string) (*Folder, error) { return nil, nil }

func (d *DriverMock) NewHost(ref *types.ManagedObjectReference) *Host { return nil }

func (d *DriverMock) FindHost(name string) (*Host, error) { return nil, nil }

func (d *DriverMock) NewNetwork(ref *types.ManagedObjectReference) *Network { return nil }

func (d *DriverMock) FindNetwork(name string) (*Network, error) { return nil, nil }

func (d *DriverMock) FindNetworks(name string) ([]*Network, error) { return nil, nil }

func (d *DriverMock) NewResourcePool(ref *types.ManagedObjectReference) *ResourcePool { return nil }

func (d *DriverMock) FindResourcePool(cluster string, host string, name string) (*ResourcePool, error) {
	return nil, nil
}

func (d *DriverMock) FindContentLibraryByName(name string) (*Library, error) { return nil, nil }

func (d *DriverMock) FindContentLibraryItem(libraryId string, name string) (*library.Item, error) {
	return nil, nil
}

func (d *DriverMock) FindContentLibraryFileDatastorePath(isoPath string) (string, error) {
	return "", nil
}
