package useragent

import (
	"testing"
)

func TestUserAgent(t *testing.T) {
	projectURL = "https://packer-test.com"
	rt = "go5.0"
	goos = "linux"
	goarch = "amd64"

	act := String("1.2.3")

	exp := "Packer/1.2.3 (+https://packer-test.com; go5.0; linux/amd64)"
	if exp != act {
		t.Errorf("expected %q to be %q", act, exp)
	}
}
