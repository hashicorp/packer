package plugin

import (
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"testing"
)

type helperBuilder byte

func (helperBuilder) Prepare(...interface{}) error {
	return nil
}

func (helperBuilder) Run(packer.Ui, packer.Hook, packer.Cache) (packer.Artifact, error) {
	return nil, nil
}

func (helperBuilder) Cancel() {}

func TestBuilder_NoExist(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.Builder()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilder_Good(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: helperProcess("builder")})
	defer c.Kill()

	_, err := c.Builder()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
