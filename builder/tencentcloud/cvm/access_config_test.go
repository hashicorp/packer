package cvm

import (
	"testing"
)

func TestTencentCloudAccessConfig_Prepare(t *testing.T) {
	cf := TencentCloudAccessConfig{
		SecretId:  "secret-id",
		SecretKey: "secret-key",
	}

	if err := cf.Prepare(nil); err == nil {
		t.Fatal("should raise error: region not set")
	}

	cf.Region = "ap-guangzhou"
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't raise error: %v", err)
	}

	cf.Region = "unknown-region"
	if err := cf.Prepare(nil); err == nil {
		t.Fatal("should raise error: unknown region")
	}

	cf.SkipValidation = true
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't raise error: %v", err)
	}
}
