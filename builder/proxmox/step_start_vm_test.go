package proxmox

import (
	"fmt"
	"testing"
	"context"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type startedVMCleanerMock struct {
	stopVm   func() (string, error)
	deleteVm func() (string, error)
}

func (m startedVMCleanerMock) StopVm(*proxmox.VmRef) (string, error) {
	return m.stopVm()
}
func (m startedVMCleanerMock) DeleteVm(*proxmox.VmRef) (string, error) {
	return m.deleteVm()
}

var _ startedVMCleaner = &startedVMCleanerMock{}

func TestCleanupStartVM(t *testing.T) {
	cs := []struct {
		name               string
		setVmRef           bool
		setSuccess         bool
		stopVMErr          error
		expectCallStopVM   bool
		deleteVMErr        error
		expectCallDeleteVM bool
	}{
		{
			name:             "when vmRef state is not set, nothing should happen",
			setVmRef:         false,
			expectCallStopVM: false,
		},
		{
			name:             "when success state is set, nothing should happen",
			setVmRef:         true,
			setSuccess:       true,
			expectCallStopVM: false,
		},
		{
			name:               "when not successful, vm should be stopped and deleted",
			setVmRef:           true,
			setSuccess:         false,
			expectCallStopVM:   true,
			expectCallDeleteVM: true,
		},
		{
			name:               "if stopping fails, DeleteVm should not be called",
			setVmRef:           true,
			setSuccess:         false,
			expectCallStopVM:   true,
			stopVMErr:          fmt.Errorf("some error"),
			expectCallDeleteVM: false,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			var stopWasCalled, deleteWasCalled bool

			cleaner := startedVMCleanerMock{
				stopVm: func() (string, error) {
					if !c.expectCallStopVM {
						t.Error("Did not expect StopVm to be called")
					}

					stopWasCalled = true
					return "", c.stopVMErr
				},
				deleteVm: func() (string, error) {
					if !c.expectCallDeleteVM {
						t.Error("Did not expect DeleteVm to be called")
					}

					deleteWasCalled = true
					return "", c.deleteVMErr
				},
			}

			state := new(multistep.BasicStateBag)
			state.Put("ui", packer.TestUi(t))
			state.Put("proxmoxClient", cleaner)
			if c.setVmRef {
				state.Put("vmRef", proxmox.NewVmRef(1))
			}
			if c.setSuccess {
				state.Put("success", "true")
			}

			step := stepStartVM{}
			step.Cleanup(state)

			if c.expectCallStopVM && !stopWasCalled {
				t.Error("Expected StopVm to be called, but it wasn't")
			}
			if c.expectCallDeleteVM && !deleteWasCalled {
				t.Error("Expected DeleteVm to be called, but it wasn't")
			}
		})
	}
}


func TestGetNextId(t *testing.T) {

	// Defines the users and their view of the clusters resources.
	// Assume three users with limited access to the clusters resources:
	//  - User1 owns VM 101 and 103
	//  - User2 owns VM 102
	//  - User3 does not own any resources yet
	// For instance, User2 does not know that User1 owns VM 103, however, User2 can
	// assume that other users created VMs by querying the global nextId, which is
	// the suggested next id to use for a new VM.

	resourcesUser1 := []Config { {
		VMID:			101,
		Pool:			"user1",
	}, {
		VMID:			103,
		Pool:			"user1",
	} }

	resourcesUser2 := []Config { {
		VMID:			102,
		Pool:			"user2",
	} }

	//resourcesUser3 := []Config{}

	// Each test case simulations creation of a new VM for a user (User1/User2/User3)
	// Thus, not all cluster resources visible in every run, e.g., in the third
	// run User3 is not able to see any resources, but can query the global nextId.

	testCases := []struct {
		name			string
		clusterResources	[]Config

		// set new VM id manually,
		// mutually exclusive with clusterResources
		manualVmId		int
	}{
		{
			name:			"Because User1 created the latest VM with the highest id, vmRef state of User1s new VM should be set to nextId",
			clusterResources:	resourcesUser1,
		},
		{
			name:			"Because User2s maxId is not the most recent vm, vmRef state of User2s new VM should be set to nextId",
			clusterResources:	resourcesUser2,
		},
		{
			name:			"when custom vmId is specified, vmRef state should contain this vmId",
			manualVmId:	199,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// with the set of users and cluster resources (vms 101,102,103),
			// the global next id (expected next id) is 104 for all users
			var globalNextId int = 104

			idMock := vmIdMock {
				// loops through the resources for this testcase/user
				// and simulates the PVE api json response
				getVmList: func() (map[string]interface{}) {
					var vmMap map[string]interface{}
					var data []interface{}
					for _, r := range(testCase.clusterResources) {
						vm := map[string]interface{}{
							"vmid": float64(r.VMID),
							"pool": r.Pool,
						}
						data = append(data, vm)
					}
					vmMap = map[string]interface{}{ "data":data }
					return vmMap
				},
				getNextID: func(currentId int) (int) {
					return globalNextId
				},
			}

			state := new(multistep.BasicStateBag)
			state.Put("ui", packer.TestUi(t))
			state.Put("proxmoxClient", idMock)

			// put an empty config by default to use next free id
			state.Put("config", &Config{})

			if testCase.manualVmId != 0 {
				// define a manual id for thew new VM
				state.Put("config", &Config {VMID: testCase.manualVmId})
			}

			// simulate the start of a new VM
			step := stepStartVM{}
			step.Run(context.TODO(), state)

			// receive the vmRef for the new VM
			// and fetch the modified config state
			vmRef := state.Get("vmRef").(*proxmox.VmRef)
			config := state.Get("config").(*Config)

			// the VMID in the config should be the same as the id in the vmRef
			if config.VMID != vmRef.VmId() {
				t.Error("Expected id of the new vmRef to be equal to the VMID in the config")
			}

			if testCase.manualVmId == 0 {
				// If there was no preference for an id, the automatically chosen global next id is used.
				// The next VM id should always be greater or equal to the global next id.
				if vmRef.VmId() < globalNextId {
					t.Error("Expected next VM id to be greater or equal to the global nextId")
				}
			} else {
				// if a manual id is used, this should be the next id
				if vmRef.VmId() != testCase.manualVmId {
					t.Error("Expected id of the new vmRef to be equal to the manually specified id")
				}
			}
		})
	}

}
