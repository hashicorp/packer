package driver

import (
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type DatastoreMock struct {
	FileExistsCalled    bool
	MakeDirectoryCalled bool
	UploadFileCalled    bool
}

func (ds *DatastoreMock) Info(params ...string) (*mo.Datastore, error) {
	return nil, nil
}

func (ds *DatastoreMock) FileExists(path string) bool {
	ds.FileExistsCalled = true
	return false
}

func (ds *DatastoreMock) Name() string {
	return "datastore-mock"
}

func (ds *DatastoreMock) Reference() types.ManagedObjectReference {
	return types.ManagedObjectReference{}
}

func (ds *DatastoreMock) ResolvePath(path string) string {
	return ""
}

func (ds *DatastoreMock) UploadFile(src, dst, host string, setHost bool) error {
	ds.UploadFileCalled = true
	return nil
}

func (ds *DatastoreMock) Delete(path string) error {
	return nil
}

func (ds *DatastoreMock) MakeDirectory(path string) error {
	ds.MakeDirectoryCalled = true
	return nil
}
