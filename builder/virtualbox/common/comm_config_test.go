package common

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

func TestCommConfigPrepare_BackwardsCompatibility(t *testing.T) {
	var c *CommConfig
	hostPortMin := 1234
	hostPortMax := 4321
	skipNatMapping := true

	c = testCommConfig()
	c.SSHHostPortMin = hostPortMin
	c.SSHHostPortMax = hostPortMax
	c.SSHSkipNatMapping = skipNatMapping

	err := c.Prepare(interpolate.NewContext())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if c.HostPortMin != hostPortMin {
		t.Fatalf("HostPortMin should be %d for backwards compatibility, but it was %d", hostPortMin, c.HostPortMin)
	}

	if c.HostPortMax != hostPortMax {
		t.Fatalf("HostPortMax should be %d for backwards compatibility, but it was %d", hostPortMax, c.HostPortMax)
	}

	if c.SkipNatMapping != skipNatMapping {
		t.Fatalf("SkipNatMapping should be %t for backwards compatibility, but it was %t", skipNatMapping, c.SkipNatMapping)
	}
}

const testPem = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAxd4iamvrwRJvtNDGQSIbNvvIQN8imXTRWlRY62EvKov60vqu
hh+rDzFYAIIzlmrJopvOe0clqmi3mIP9dtkjPFrYflq52a2CF5q+BdwsJXuRHbJW
LmStZUwW1khSz93DhvhmK50nIaczW63u4EO/jJb3xj+wxR1Nkk9bxi3DDsYFt8SN
AzYx9kjlEYQ/+sI4/ATfmdV9h78SVotjScupd9KFzzi76gWq9gwyCBLRynTUWlyD
2UOfJRkOvhN6/jKzvYfVVwjPSfA9IMuooHdScmC4F6KBKJl/zf/zETM0XyzIDNmH
uOPbCiljq2WoRM+rY6ET84EO0kVXbfx8uxUsqQIDAQABAoIBAQCkPj9TF0IagbM3
5BSs/CKbAWS4dH/D4bPlxx4IRCNirc8GUg+MRb04Xz0tLuajdQDqeWpr6iLZ0RKV
BvreLF+TOdV7DNQ4XE4gSdJyCtCaTHeort/aordL3l0WgfI7mVk0L/yfN1PEG4YG
E9q1TYcyrB3/8d5JwIkjabxERLglCcP+geOEJp+QijbvFIaZR/n2irlKW4gSy6ko
9B0fgUnhkHysSg49ChHQBPQ+o5BbpuLrPDFMiTPTPhdfsvGGcyCGeqfBA56oHcSF
K02Fg8OM+Bd1lb48LAN9nWWY4WbwV+9bkN3Ym8hO4c3a/Dxf2N7LtAQqWZzFjvM3
/AaDvAgBAoGBAPLD+Xn1IYQPMB2XXCXfOuJewRY7RzoVWvMffJPDfm16O7wOiW5+
2FmvxUDayk4PZy6wQMzGeGKnhcMMZTyaq2g/QtGfrvy7q1Lw2fB1VFlVblvqhoJa
nMJojjC4zgjBkXMHsRLeTmgUKyGs+fdFbfI6uejBnnf+eMVUMIdJ+6I9AoGBANCn
kWO9640dttyXURxNJ3lBr2H3dJOkmD6XS+u+LWqCSKQe691Y/fZ/ZL0Oc4Mhy7I6
hsy3kDQ5k2V0fkaNODQIFJvUqXw2pMewUk8hHc9403f4fe9cPrL12rQ8WlQw4yoC
v2B61vNczCCUDtGxlAaw8jzSRaSI5s6ax3K7enbdAoGBAJB1WYDfA2CoAQO6y9Sl
b07A/7kQ8SN5DbPaqrDrBdJziBQxukoMJQXJeGFNUFD/DXFU5Fp2R7C86vXT7HIR
v6m66zH+CYzOx/YE6EsUJms6UP9VIVF0Rg/RU7teXQwM01ZV32LQ8mswhTH20o/3
uqMHmxUMEhZpUMhrfq0isyApAoGAe1UxGTXfj9AqkIVYylPIq2HqGww7+jFmVEj1
9Wi6S6Sq72ffnzzFEPkIQL/UA4TsdHMnzsYKFPSbbXLIWUeMGyVTmTDA5c0e5XIR
lPhMOKCAzv8w4VUzMnEkTzkFY5JqFCD/ojW57KvDdNZPVB+VEcdxyAW6aKELXMAc
eHLc1nkCgYEApm/motCTPN32nINZ+Vvywbv64ZD+gtpeMNP3CLrbe1X9O+H52AXa
1jCoOldWR8i2bs2NVPcKZgdo6fFULqE4dBX7Te/uYEIuuZhYLNzRO1IKU/YaqsXG
3bfQ8hKYcSnTfE0gPtLDnqCIxTocaGLSHeG3TH9fTw+dA8FvWpUztI4=
-----END RSA PRIVATE KEY-----
`
