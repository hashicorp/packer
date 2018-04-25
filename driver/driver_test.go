package driver

import (
	"fmt"
	"testing"
	"time"
	"math/rand"
	"context"
)

// Defines whether acceptance tests should be run
const TestHostName = "esxi-1.vsphere65.test"

func newTestDriver(t *testing.T) *Driver {
	d, err := NewDriver(context.TODO(), &ConnectConfig{
		VCenterServer:      "vcenter.vsphere65.test",
		Username:           "root",
		Password:           "jetbrains",
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
