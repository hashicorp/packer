package plugin

import (
	"context"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer/packer"
)

type helperPostProcessor byte

func (helperPostProcessor) Configure(...interface{}) error {
	return nil
}

func (helperPostProcessor) PostProcess(context.Context, packer.Ui, packer.Artifact) (packer.Artifact, bool, error) {
	return nil, false, nil
}

func TestPostProcessor_NoExist(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.PostProcessor()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestPostProcessor_Good(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: helperProcess("post-processor")})
	defer c.Kill()

	_, err := c.PostProcessor()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
