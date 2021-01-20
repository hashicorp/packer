package plugin

import (
	"os/exec"
	"testing"
)

func TestDatasource_NoExist(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.Datasource()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestDatasource_Good(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: helperProcess("datasource")})
	defer c.Kill()

	_, err := c.Datasource()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
