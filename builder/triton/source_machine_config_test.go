package triton

import (
	"testing"
)

func TestSourceMachineConfig_Prepare(t *testing.T) {
	sc := testSourceMachineConfig(t)
	errs := sc.Prepare(nil)
	if errs != nil {
		t.Fatalf("should not error: %#v", sc)
	}

	sc = testSourceMachineConfig(t)
	sc.MachineName = ""
	errs = sc.Prepare(nil)
	if errs == nil {
		t.Fatalf("should error: %#v", sc)
	}

	sc = testSourceMachineConfig(t)
	sc.MachinePackage = ""
	errs = sc.Prepare(nil)
	if errs == nil {
		t.Fatalf("should error: %#v", sc)
	}

	sc = testSourceMachineConfig(t)
	sc.MachineImage = ""
	errs = sc.Prepare(nil)
	if errs == nil {
		t.Fatalf("should error: %#v", sc)
	}
}

func testSourceMachineConfig(t *testing.T) SourceMachineConfig {
	return SourceMachineConfig{
		MachineName:    "test-machine",
		MachinePackage: "test-package",
		MachineImage:   "test-image",
		MachineNetworks: []string{
			"test-network-1",
			"test-network-2",
		},
		MachineMetadata: map[string]string{
			"test-metadata-key1": "test-metadata-value1",
			"test-metadata-key2": "test-metadata-value2",
			"test-metadata-key3": "test-metadata-value3",
		},
		MachineTags: map[string]string{
			"test-tags-key1": "test-tags-value1",
			"test-tags-key2": "test-tags-value2",
			"test-tags-key3": "test-tags-value3",
		},
		MachineFirewallEnabled: false,
	}
}
