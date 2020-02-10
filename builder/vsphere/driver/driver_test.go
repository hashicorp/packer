package driver

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

// Defines whether acceptance tests should be run
const TestHostName = "esxi-1.vsphere65.test"

func newTestDriver(t *testing.T) *Driver {
	username := os.Getenv("VSPHERE_USERNAME")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("VSPHERE_PASSWORD")
	if password == "" {
		password = "jetbrains"
	}

	d, err := NewDriver(&ConnectConfig{
		VCenterServer:      "vcenter.vsphere65.test",
		Username:           username,
		Password:           password,
		InsecureConnection: true,
	})
	if err != nil {
		t.Fatalf("Cannot connect: %v", err)
	}
	return d
}

func newVMName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return fmt.Sprintf("test-%v", rand.Intn(1000))
}
