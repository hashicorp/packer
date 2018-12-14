package tencent

import (
	"os"
	"testing"
)

func TestGetRequiredEnvVars(t *testing.T) {
	k1 := CurrentTimeStamp()
	v1 := CurrentTimeStamp()
	err1 := os.Setenv(k1, v1)
	if err1 != nil {
		t.Fatalf("GetRequiredEnvVar: Failed to set OS environment variable")
	}
	requiredEnvVar := map[string]string{k1: ""}
	GetRequiredEnvVars(requiredEnvVar)
	v2 := requiredEnvVar[k1]
	if v1 != v2 {
		t.Fatalf("GetRequiredEnvVar failed, %s != %s", v1, v2)
	}
	os.Unsetenv(k1)
}

func TestGetEnvVar(t *testing.T) {
	path := GetEnvVar("PATH")
	if  path == "" {
		t.Fatalf("GetEnvVar failed, path: %s", path)
	}
}
