package driver

import (
	"os"
	"fmt"
	"testing"
	"time"
	"math/rand"
)

// Defines whether acceptance tests should be run
const TestEnvVar = "VSPHERE_DRIVER_ACC"
const hostName = "esxi-1.vsphere55.test"

func newTestDriver(t *testing.T) *Driver {
	d, err := NewDriver(&ConnectConfig{
		VCenterServer:      "vcenter.vsphere55.test",
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

func initDriverAcceptanceTest(t *testing.T) {
	// We only run acceptance tests if an env var is set because they're
	// slow and require outside configuration.
	if os.Getenv(TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			TestEnvVar))
	}

	// We require verbose mode so that the user knows what is going on.
	if !testing.Verbose() {
		t.Fatal("Acceptance tests must be run with the -v flag on tests")
	}
}
