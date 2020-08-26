package driver

import (
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vim25/types"
)

type DriverMock struct {
	FindDatastoreCalled bool
	DatastoreMock       *DatastoreMock
}

func NewDriverMock() *DriverMock {
	return new(DriverMock)
}

func (d *DriverMock) FindDatastore(name string, host string) (Datastore, error) {
	d.FindDatastoreCalled = true
	if d.DatastoreMock == nil {
		d.DatastoreMock = new(DatastoreMock)
	}
	return d.DatastoreMock, nil
}

func (d *DriverMock) NewVM(ref *types.ManagedObjectReference) *VirtualMachine {
	return nil
}

func (d *DriverMock) FindVM(name string) (*VirtualMachine, error) {
	return nil, nil
}

func (d *DriverMock) FindCluster(name string) (*Cluster, error) {
	return nil, nil
}

func (d *DriverMock) PreCleanVM(ui packer.Ui, vmPath string, force bool) error {
	return nil
}

func (d *DriverMock) CreateVM(config *CreateConfig) (*VirtualMachine, error) { return nil, nil }

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
