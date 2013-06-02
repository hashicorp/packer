package plugin

import (
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	process := helperProcess("mock")
	c := NewClient(process)
	defer c.Kill()

	// Test that it parses the proper address
	addr, err := c.Start()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	if addr != ":1234" {
		t.Fatalf("incorrect addr %s", addr)
	}

	// Test that it exits properly if killed
	c.Kill()

	if process.ProcessState == nil {
		t.Fatal("should have process state")
	}

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_Start_Timeout(t *testing.T) {
	c := NewClient(helperProcess("start-timeout"))
	defer c.Kill()

	// Set a shorter timeout
	c.StartTimeout = 50 * time.Millisecond

	_, err := c.Start()
	if err == nil {
		t.Fatal("err should not be nil")
	}
}
