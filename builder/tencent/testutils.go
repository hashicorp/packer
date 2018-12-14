package tencent

import (
	"testing"
)

// CheckConfigHasErrors checks that a test has errors
func CheckConfigHasErrors(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should error")
	}
}

// CheckConfigIsOk checks that a test has no errors
func CheckConfigIsOk(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}

// Removes a particular string from a string array
func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func NewDefaultConfig() Config {
	requiredEnvVars := map[string]string{
		CRegion:        "ap-singapore",
		CPlacementZone: "ap-singapore-1",
	}
	GetRequiredEnvVars(requiredEnvVars)
	return Config{
		ImageName: CurrentTimeStamp(),
		Placement: CVMPlacement{Zone: requiredEnvVars[CPlacementZone]},
		Region:    requiredEnvVars[CRegion],
		SecretID:  requiredEnvVars[CSecretId],
		SecretKey: requiredEnvVars[CSecretKey],
		Timeout:   300000,
		Url:       CCVMUrlSiliconValley,
	}
}
