package common

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/builder/vsphere/driver"
)

func TestStepRemoveFloppy_Run(t *testing.T) {
	tc := []struct {
		name               string
		uploadedPath       string
		step               *StepRemoveFloppy
		expectedAction     multistep.StepAction
		vmMock             *driver.VirtualMachineMock
		expectedVmMock     *driver.VirtualMachineMock
		driverMock         *driver.DriverMock
		expectedDriverMock *driver.DriverMock
		dsMock             *driver.DatastoreMock
		expectedDsMock     *driver.DatastoreMock
		fail               bool
		errMessage         string
	}{
		{
			name:         "Remove floppy drives and images",
			uploadedPath: "vm/dir/packer-tmp-created-floppy.flp",
			step: &StepRemoveFloppy{
				Datastore: "datastore",
				Host:      "host",
			},
			expectedAction: multistep.ActionContinue,
			vmMock:         new(driver.VirtualMachineMock),
			expectedVmMock: &driver.VirtualMachineMock{
				FloppyDevicesCalled:   true,
				RemoveDeviceCalled:    true,
				RemoveDeviceKeepFiles: true,
			},
			driverMock: new(driver.DriverMock),
			expectedDriverMock: &driver.DriverMock{
				FindDatastoreCalled: true,
				FindDatastoreName:   "datastore",
				FindDatastoreHost:   "host",
			},
			dsMock: new(driver.DatastoreMock),
			expectedDsMock: &driver.DatastoreMock{
				DeleteCalled: true,
				DeletePath:   "vm/dir/packer-tmp-created-floppy.flp",
			},
			fail: false,
		},
		{
			name:           "No floppy image to remove",
			step:           &StepRemoveFloppy{},
			expectedAction: multistep.ActionContinue,
			vmMock:         new(driver.VirtualMachineMock),
			expectedVmMock: &driver.VirtualMachineMock{
				FloppyDevicesCalled:   true,
				RemoveDeviceCalled:    true,
				RemoveDeviceKeepFiles: true,
			},
			driverMock:         new(driver.DriverMock),
			expectedDriverMock: new(driver.DriverMock),
			dsMock:             new(driver.DatastoreMock),
			expectedDsMock:     new(driver.DatastoreMock),
			fail:               false,
		},
		{
			name:           "Fail to find floppy devices",
			step:           &StepRemoveFloppy{},
			expectedAction: multistep.ActionHalt,
			vmMock: &driver.VirtualMachineMock{
				FloppyDevicesErr: fmt.Errorf("failed to find floppy devices"),
			},
			expectedVmMock: &driver.VirtualMachineMock{
				FloppyDevicesCalled: true,
			},
			driverMock:         new(driver.DriverMock),
			expectedDriverMock: new(driver.DriverMock),
			dsMock:             new(driver.DatastoreMock),
			expectedDsMock:     new(driver.DatastoreMock),
			fail:               true,
			errMessage:         "failed to find floppy devices",
		},
		{
			name:           "Fail to remove floppy devices",
			step:           &StepRemoveFloppy{},
			expectedAction: multistep.ActionHalt,
			vmMock: &driver.VirtualMachineMock{
				RemoveDeviceErr: fmt.Errorf("failed to remove device"),
			},
			expectedVmMock: &driver.VirtualMachineMock{
				FloppyDevicesCalled:   true,
				RemoveDeviceCalled:    true,
				RemoveDeviceKeepFiles: true,
			},
			driverMock:         new(driver.DriverMock),
			expectedDriverMock: new(driver.DriverMock),
			dsMock:             new(driver.DatastoreMock),
			expectedDsMock:     new(driver.DatastoreMock),
			fail:               true,
			errMessage:         "failed to remove device",
		},
		{
			name:         "Fail to find datastore",
			uploadedPath: "vm/dir/packer-tmp-created-floppy.flp",
			step: &StepRemoveFloppy{
				Datastore: "datastore",
				Host:      "host",
			},
			expectedAction: multistep.ActionHalt,
			vmMock:         new(driver.VirtualMachineMock),
			expectedVmMock: &driver.VirtualMachineMock{
				FloppyDevicesCalled:   true,
				RemoveDeviceCalled:    true,
				RemoveDeviceKeepFiles: true,
			},
			driverMock: &driver.DriverMock{
				FindDatastoreErr: fmt.Errorf("failed to find datastore"),
			},
			expectedDriverMock: &driver.DriverMock{
				FindDatastoreCalled: true,
				FindDatastoreName:   "datastore",
				FindDatastoreHost:   "host",
			},
			dsMock:         new(driver.DatastoreMock),
			expectedDsMock: new(driver.DatastoreMock),
			fail:           true,
			errMessage:     "failed to find datastore",
		},
		{
			name:         "Fail to delete floppy image",
			uploadedPath: "vm/dir/packer-tmp-created-floppy.flp",
			step: &StepRemoveFloppy{
				Datastore: "datastore",
				Host:      "host",
			},
			expectedAction: multistep.ActionHalt,
			vmMock:         new(driver.VirtualMachineMock),
			expectedVmMock: &driver.VirtualMachineMock{
				FloppyDevicesCalled:   true,
				RemoveDeviceCalled:    true,
				RemoveDeviceKeepFiles: true,
			},
			driverMock: new(driver.DriverMock),
			expectedDriverMock: &driver.DriverMock{
				FindDatastoreCalled: true,
				FindDatastoreName:   "datastore",
				FindDatastoreHost:   "host",
			},
			dsMock: &driver.DatastoreMock{
				DeleteErr: fmt.Errorf("failed to delete floppy"),
			},
			expectedDsMock: &driver.DatastoreMock{
				DeleteCalled: true,
				DeletePath:   "vm/dir/packer-tmp-created-floppy.flp",
			},
			fail:       true,
			errMessage: "failed to delete floppy",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			state := basicStateBag(nil)
			state.Put("vm", c.vmMock)
			c.driverMock.DatastoreMock = c.dsMock
			state.Put("driver", c.driverMock)

			if c.uploadedPath != "" {
				state.Put("uploaded_floppy_path", c.uploadedPath)
			}

			if action := c.step.Run(context.TODO(), state); action != c.expectedAction {
				t.Fatalf("unexpected action %v", action)
			}
			err, ok := state.Get("error").(error)
			if ok {
				if err.Error() != c.errMessage {
					t.Fatalf("unexpected error %s", err.Error())
				}
			} else {
				if c.fail {
					t.Fatalf("expected to fail but it didn't")
				}
			}

			if !c.fail {
				if _, ok := state.GetOk("uploaded_floppy_path"); ok {
					t.Fatalf("uploaded_floppy_path should not be in state")
				}
			}

			if diff := cmp.Diff(c.vmMock, c.expectedVmMock,
				cmpopts.IgnoreInterfaces(struct{ error }{})); diff != "" {
				t.Fatalf("unexpected VirtualMachine calls: %s", diff)
			}
			c.expectedDriverMock.DatastoreMock = c.expectedDsMock
			if diff := cmp.Diff(c.driverMock, c.expectedDriverMock,
				cmpopts.IgnoreInterfaces(struct{ error }{})); diff != "" {
				t.Fatalf("unexpected Driver calls: %s", diff)
			}
			if diff := cmp.Diff(c.dsMock, c.expectedDsMock,
				cmpopts.IgnoreInterfaces(struct{ error }{})); diff != "" {
				t.Fatalf("unexpected Datastore calls: %s", diff)
			}
		})
	}
}
