package plugin

import (
	"os/exec"
	"testing"
)

func TestProvisioner_NoExist(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.Provisioner()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisioner_Good(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: helperProcess("provisioner")})
	defer c.Kill()

	_, err := c.Provisioner()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
