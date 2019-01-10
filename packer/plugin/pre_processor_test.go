package plugin

import (
	"os/exec"
	"testing"

	"github.com/hashicorp/packer/packer"
)

type helperPreProcessor byte

func (helperPreProcessor) Configure(...interface{}) error {
	return nil
}

func (helperPreProcessor) PreProcess(packer.Ui) error {
	return nil
}

func TestPreProcessor_NoExist(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.PreProcessor()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestPreProcessor_Good(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: helperProcess("post-processor")})
	defer c.Kill()

	_, err := c.PreProcessor()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
