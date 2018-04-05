package useragent

import (
	"testing"
)

func TestUserAgent(t *testing.T) {
	projectURL = "https://packer-test.com"
	rt = "go5.0"
	versionFunc = func() string { return "1.2.3" }

	act := String()

	exp := "Packer/1.2.3 (+https://packer-test.com; go5.0)"
	if exp != act {
		t.Errorf("expected %q to be %q", act, exp)
	}
}
