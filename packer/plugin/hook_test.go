package plugin

import (
	"os/exec"
	"testing"
)

func TestHook_NoExist(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.Hook()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestHook_Good(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: helperProcess("hook")})
	defer c.Kill()

	_, err := c.Hook()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
