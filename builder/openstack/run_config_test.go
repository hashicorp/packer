package openstack

import (
	"os"
	"testing"

	"github.com/hashicorp/packer/helper/communicator"
)

func init() {
	// Clear out the openstack env vars so they don't
	// affect our tests.
	os.Setenv("SDK_USERNAME", "")
	os.Setenv("SDK_PASSWORD", "")
	os.Setenv("SDK_PROVIDER", "")
}

func testRunConfig() *RunConfig {
	return &RunConfig{
		SourceImage: "abcd",
		Flavor:      "m1.small",

		Comm: communicator.Config{
			SSHUsername: "foo",
		},
	}
}

func TestRunConfigPrepare(t *testing.T) {
	c := testRunConfig()
	err := c.Prepare(nil)
	if len(err) > 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_InstanceType(t *testing.T) {
	c := testRunConfig()
	c.Flavor = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SourceImage(t *testing.T) {
	c := testRunConfig()
	c.SourceImage = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_SSHPort(t *testing.T) {
	c := testRunConfig()
	c.Comm.SSHPort = 0
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHPort != 22 {
		t.Fatalf("invalid value: %d", c.Comm.SSHPort)
	}

	c.Comm.SSHPort = 44
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHPort != 44 {
		t.Fatalf("invalid value: %d", c.Comm.SSHPort)
	}
}

func TestRunConfigPrepare_BlockStorage(t *testing.T) {
	c := testRunConfig()
	c.UseBlockStorageVolume = true
	c.VolumeType = "fast"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
	if c.VolumeType != "fast" {
		t.Fatalf("invalid value: %s", c.VolumeType)
	}

	c.AvailabilityZone = "RegionTwo"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.VolumeAvailabilityZone != "RegionTwo" {
		t.Fatalf("invalid value: %s", c.VolumeAvailabilityZone)
	}

	c.VolumeAvailabilityZone = "RegionOne"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.VolumeAvailabilityZone != "RegionOne" {
		t.Fatalf("invalid value: %s", c.VolumeAvailabilityZone)
	}

	c.VolumeName = "PackerVolume"
	if c.VolumeName != "PackerVolume" {
		t.Fatalf("invalid value: %s", c.VolumeName)
	}
}

func TestRunConfigPrepare_FloatingIPPoolCompat(t *testing.T) {
	c := testRunConfig()
	c.FloatingIPPool = "uuid1"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.FloatingIPNetwork != "uuid1" {
		t.Fatalf("invalid value: %s", c.FloatingIPNetwork)
	}

	c.FloatingIPNetwork = "uuid2"
	c.FloatingIPPool = "uuid3"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.FloatingIPNetwork != "uuid2" {
		t.Fatalf("invalid value: %s", c.FloatingIPNetwork)
	}
}
