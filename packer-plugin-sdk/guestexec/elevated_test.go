package guestexec

import (
	"regexp"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"inline": []interface{}{"foo", "bar"},
	}
}

func TestProvisioner_GenerateElevatedRunner(t *testing.T) {

	// Non-elevated
	config := testConfig()
	p := new(packer.MockProvisioner)
	p.Prepare(config)
	comm := new(packer.MockCommunicator)
	p.ProvCommunicator = comm
	path, err := GenerateElevatedRunner("whoami", p)

	if err != nil {
		t.Fatalf("Did not expect error: %s", err.Error())
	}

	if comm.UploadCalled != true {
		t.Fatalf("Should have uploaded file")
	}

	matched, _ := regexp.MatchString("C:/Windows/Temp/packer-elevated-shell.*", path)
	if !matched {
		t.Fatalf("Got unexpected file: %s", path)
	}
}
