package triton

import (
	"testing"
)

func TestAccessConfig_Prepare(t *testing.T) {
	ac := testAccessConfig(t)
	errs := ac.Prepare(nil)
	if errs != nil {
		t.Fatal("should not error")
	}

	ac = testAccessConfig(t)
	ac.Account = ""
	errs = ac.Prepare(nil)
	if errs == nil {
		t.Fatal("should error")
	}

	ac = testAccessConfig(t)
	ac.KeyID = ""
	errs = ac.Prepare(nil)
	if errs == nil {
		t.Fatal("should error")
	}

	ac = testAccessConfig(t)
	ac.KeyMaterial = ""
	errs = ac.Prepare(nil)
	if errs == nil {
		t.Fatal("should error")
	}
}

func testAccessConfig(t *testing.T) AccessConfig {
	return AccessConfig{
		Endpoint: "test-endpoint",
		Account:  "test-account",
		KeyID:    "test-id",
		KeyMaterial:  "test-private-key",
	}
}
