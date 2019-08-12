package jdcloud

import (
	"testing"
)

func TestJDCloudCredentialConfig_Prepare(t *testing.T) {

	creds := &JDCloudCredentialConfig{}

	if err := creds.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when there's nothing set")
	}

	creds.AccessKey = "abc"
	if err := creds.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when theres no Secret key")
	}

	creds.SecretKey = "123"
	if err := creds.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when theres no Az and region")
	}

	creds.RegionId = "cn-west-1"
	creds.Az = "cn-north-1c"
	if err := creds.Prepare(nil); err == nil {
		t.Fatalf("Test shouldn't pass when region_id illegal")
	}
	creds.RegionId = "cn-north-1"
	if err := creds.Prepare(nil); err != nil {
		t.Fatalf("Test shouldn't fail...")
	}
}
