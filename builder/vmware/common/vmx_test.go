package common

import "testing"

func TestParseVMX(t *testing.T) {
	contents := `
.encoding = "UTF-8"
config.version = "8"
scsi0:0.virtualSSD = 1
`

	results := ParseVMX(contents)
	if len(results) != 3 {
		t.Fatalf("not correct number of results: %d", len(results))
	}

	if results[".encoding"] != "UTF-8" {
		t.Errorf("invalid .encoding: %s", results[".encoding"])
	}

	if results["config.version"] != "8" {
		t.Errorf("invalid config.version: %s", results["config.version"])
	}

	if results["scsi0:0.virtualssd"] != "1" {
		t.Errorf("invalid scsi0:0.virtualssd: %s", results["scsi0:0.virtualssd"])
	}
}

func TestEncodeVMX(t *testing.T) {
	contents := map[string]string{
		".encoding":          "UTF-8",
		"config.version":     "8",
		"scsi0:0.virtualssd": "1",
	}

	expected := `.encoding = "UTF-8"
config.version = "8"
scsi0:0.virtualSSD = 1
`

	result := EncodeVMX(contents)
	if result != expected {
		t.Errorf("invalid results: %s", result)
	}
}
