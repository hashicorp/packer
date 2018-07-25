package vmx

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/stretchr/testify/assert"
)

const (
	scsiFilename = "scsiDisk.vmdk"
	sataFilename = "sataDisk.vmdk"
	nvmeFilename = "nvmeDisk.vmdk"
	ideFilename  = "ideDisk.vmdk"
)

func TestStepCloneVMX_impl(t *testing.T) {
	var _ multistep.Step = new(StepCloneVMX)
}

func TestStepCloneVMX(t *testing.T) {
	// Setup some state
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	// Set up mock vmx file contents
	var testCloneVMX = fmt.Sprintf("scsi0:0.filename = \"%s\"\n"+
		"sata0:0.filename = \"%s\"\n"+
		"nvme0:0.filename = \"%s\"\n"+
		"ide1:0.filename = \"%s\"\n"+
		"ide0:0.filename = \"auto detect\"\n"+
		"ethernet0.connectiontype = \"nat\"\n", scsiFilename,
		sataFilename, nvmeFilename, ideFilename)

	// Set up expected mock disk file paths
	diskFilenames := []string{scsiFilename, sataFilename, ideFilename, nvmeFilename}
	var diskFullPaths []string
	for _, diskFilename := range diskFilenames {
		diskFullPaths = append(diskFullPaths, filepath.Join(td, diskFilename))
	}

	// Create the source
	sourcePath := filepath.Join(td, "source.vmx")
	if err := ioutil.WriteFile(sourcePath, []byte(testCloneVMX), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Create the dest because the mock driver won't
	destPath := filepath.Join(td, "foo.vmx")
	if err := ioutil.WriteFile(destPath, []byte(testCloneVMX), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	state := testState(t)
	step := new(StepCloneVMX)
	step.OutputDir = td
	step.Path = sourcePath
	step.VMName = "foo"

	driver := state.Get("driver").(*vmwcommon.DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test we cloned
	if !driver.CloneCalled {
		t.Fatal("should call clone")
	}

	// Test that we have our paths
	if vmxPath, ok := state.GetOk("vmx_path"); !ok {
		t.Fatal("should set vmx_path")
	} else if vmxPath != destPath {
		t.Fatalf("bad path to vmx: %#v", vmxPath)
	}

	if stateDiskPaths, ok := state.GetOk("disk_full_paths"); !ok {
		t.Fatal("should set disk_full_paths")
	} else {
		assert.ElementsMatchf(t, stateDiskPaths.([]string), diskFullPaths,
			"%s\nshould contain the same elements as:\n%s", stateDiskPaths.([]string), diskFullPaths)
	}

	// Test we got the network type
	if networkType, ok := state.GetOk("vmnetwork"); !ok {
		t.Fatal("should set vmnetwork")
	} else if networkType != "nat" {
		t.Fatalf("bad network type: %#v", networkType)
	}
}
