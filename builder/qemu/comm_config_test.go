package qemu

import (
	"io/ioutil"
	"os"
	"testing"
	
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

func testCommConfig() *CommConfig {
	return &CommConfig{
		Comm: communicator.Config{
			SSH: communicator.SSH{
				SSHUsername: "foo",
			},
		},
	}
}

func TestCommConfigPrepare(t *testing.T) {
	c := testCommConfig()
	errs := c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if c.HostPortMin != 2222 {
		t.Errorf("bad min communicator host port: %d", c.HostPortMin)
	}

	if c.HostPortMax != 4444 {
		t.Errorf("bad max communicator host port: %d", c.HostPortMax)
	}

	if c.Comm.SSHPort != 22 {
		t.Errorf("bad communicator port: %d", c.Comm.SSHPort)
	}
}

func TestCommConfigPrepare_SSHHostPort(t *testing.T) {
	var c *CommConfig
	var errs []error

	// Bad
	c = testCommConfig()
	c.HostPortMin = 1000
	c.HostPortMax = 500
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatalf("bad: %#v", errs)
	}

	// Good
	c = testCommConfig()
	c.HostPortMin = 50
	c.HostPortMax = 500
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
}

func TestCommConfigPrepare_SSHPrivateKey(t *testing.T) {
	var c *CommConfig
	var errs []error

	c = testCommConfig()
	c.Comm.SSHPrivateKeyFile = ""
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}

	c = testCommConfig()
	c.Comm.SSHPrivateKeyFile = "/i/dont/exist"
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatal("should have error")
	}

	// Test bad contents
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	if _, err := tf.Write([]byte("HELLO!")); err != nil {
		t.Fatalf("err: %s", err)
	}

	c = testCommConfig()
	c.Comm.SSHPrivateKeyFile = tf.Name()
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatal("should have error")
	}

	// Test good contents
	tf.Seek(0, 0)
	tf.Truncate(0)
	tf.Write([]byte(testPem))
	c = testCommConfig()
	c.Comm.SSHPrivateKeyFile = tf.Name()
	errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}
}
