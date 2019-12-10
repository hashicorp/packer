package driver

import (
	"log"
	"testing"
)

func TestVMAcc_create(t *testing.T) {
	testCases := []struct {
		name          string
		config        *CreateConfig
		checkFunction func(*testing.T, *VirtualMachine, *CreateConfig)
	}{
		{"MinimalConfiguration", &CreateConfig{}, createDefaultCheck},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.Host = TestHostName
			tc.config.Name = newVMName()

			d := newTestDriver(t)

			log.Printf("[DEBUG] Creating VM")
			vm, err := d.CreateVM(tc.config)
			if err != nil {
				t.Fatalf("Cannot create VM: %v", err)
			}

			defer destroyVM(t, vm, tc.config.Name)

			log.Printf("[DEBUG] Running check function")
			tc.checkFunction(t, vm, tc.config)
		})
	}
}

func createDefaultCheck(t *testing.T, vm *VirtualMachine, config *CreateConfig) {
	d := vm.driver

	// Check that the clone can be found by its name
	if _, err := d.FindVM(config.Name); err != nil {
		t.Errorf("Cannot find created vm '%v': %v", config.Name, err)
	}

	vmInfo, err := vm.Info("name", "parent", "runtime.host", "resourcePool", "datastore", "layoutEx.disk")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	if vmInfo.Name != config.Name {
		t.Errorf("Invalid VM name: expected '%v', got '%v'", config.Name, vmInfo.Name)
	}

	f := d.NewFolder(vmInfo.Parent)
	folderPath, err := f.Path()
	if err != nil {
		t.Fatalf("Cannot read folder name: %v", err)
	}
	if folderPath != "" {
		t.Errorf("Invalid folder: expected '/', got '%v'", folderPath)
	}

	h := d.NewHost(vmInfo.Runtime.Host)
	hostInfo, err := h.Info("name")
	if err != nil {
		t.Fatal("Cannot read host properties: ", err)
	}
	if hostInfo.Name != TestHostName {
		t.Errorf("Invalid host name: expected '%v', got '%v'", TestHostName, hostInfo.Name)
	}

	p := d.NewResourcePool(vmInfo.ResourcePool)
	poolPath, err := p.Path()
	if err != nil {
		t.Fatalf("Cannot read resource pool name: %v", err)
	}
	if poolPath != "" {
		t.Errorf("Invalid resource pool: expected '/', got '%v'", poolPath)
	}

	dsr := vmInfo.Datastore[0].Reference()
	ds := d.NewDatastore(&dsr)
	dsInfo, err := ds.Info("name")
	if err != nil {
		t.Fatal("Cannot read datastore properties: ", err)
	}
	if dsInfo.Name != "datastore1" {
		t.Errorf("Invalid datastore name: expected 'datastore1', got '%v'", dsInfo.Name)
	}
}
