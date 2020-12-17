package common

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/builder/vsphere/driver"
)

func TestStepRemoteUpload_Run(t *testing.T) {
	state := basicStateBag(nil)
	dsMock := driver.DatastoreMock{
		DirExistsReturn: false,
	}
	driverMock := driver.NewDriverMock()
	driverMock.DatastoreMock = &dsMock
	state.Put("driver", driverMock)
	state.Put("iso_path", "[datastore] iso/path")

	step := &StepRemoteUpload{
		Datastore:                  "datastore",
		Host:                       "host",
		SetHostForDatastoreUploads: false,
	}

	if action := step.Run(context.TODO(), state); action == multistep.ActionHalt {
		t.Fatalf("Should not halt.")
	}

	if !driverMock.FindDatastoreCalled {
		t.Fatalf("driver.FindDatastore should be called.")
	}
	if !driverMock.DatastoreMock.FileExistsCalled {
		t.Fatalf("datastore.FindDatastore should be called.")
	}
	if !driverMock.DatastoreMock.MakeDirectoryCalled {
		t.Fatalf("datastore.MakeDirectory should be called.")
	}
	if !driverMock.DatastoreMock.UploadFileCalled {
		t.Fatalf("datastore.UploadFile should be called.")
	}
	remotePath, ok := state.GetOk("iso_remote_path")
	if !ok {
		t.Fatalf("state should contain iso_remote_path")
	}
	expectedRemovePath := fmt.Sprintf("[%s] packer_cache//path", driverMock.DatastoreMock.Name())
	if remotePath != expectedRemovePath {
		t.Fatalf("iso_remote_path expected to be %s but was %s", expectedRemovePath, remotePath)
	}
}

func TestStepRemoteUpload_SkipRun(t *testing.T) {
	state := basicStateBag(nil)
	driverMock := driver.NewDriverMock()
	state.Put("driver", driverMock)

	step := &StepRemoteUpload{}

	if action := step.Run(context.TODO(), state); action == multistep.ActionHalt {
		t.Fatalf("Should not halt.")
	}

	if driverMock.FindDatastoreCalled {
		t.Fatalf("driver.FindDatastore should not be called.")
	}
	if _, ok := state.GetOk("iso_remote_path"); ok {
		t.Fatalf("state should not contain iso_remote_path")
	}
}
