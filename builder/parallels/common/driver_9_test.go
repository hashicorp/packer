package common

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestParallels9Driver_impl(t *testing.T) {
	var _ Driver = new(Parallels9Driver)
}

func TestIPAddress(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())

	d := Parallels9Driver{
		dhcpLeaseFile: tf.Name(),
	}

	// No lease should be found in an empty file
	ip, err := d.IPAddress("123456789012")
	if err == nil {
		t.Fatalf("Found IP: \"%v\". No IP should be found!\n", ip)
	}

	// The most recent lease, 10.211.55.126 should be found
	c := []byte(`
[vnic0]
10.211.55.125="1418288000,1800,001c4235240c,ff4235240c000100011c1c10e7001c4235240c"
10.211.55.126="1418288969,1800,001c4235240c,ff4235240c000100011c1c11ad001c4235240c"
10.211.55.254="1411712008,1800,001c42a51419,01001c42a51419"
`)
	ioutil.WriteFile(tf.Name(), c, 0666)
	ip, err = d.IPAddress("001C4235240c")
	if err != nil {
		t.Fatalf("Error: %v\n", err)
	}
	if ip != "10.211.55.126" {
		t.Fatalf("Should have found 10.211.55.126, not %s!\n", ip)
	}

	// The most recent lease, 10.211.55.124 should be found
	c = []byte(`[vnic0]
10.211.55.124="1418288969,1800,001c4235240c,ff4235240c000100011c1c11ad001c4235240c"
10.211.55.125="1418288000,1800,001c4235240c,ff4235240c000100011c1c10e7001c4235240c"
10.211.55.254="1411712008,1800,001c42a51419,01001c42a51419"
`)
	ioutil.WriteFile(tf.Name(), c, 0666)
	ip, err = d.IPAddress("001c4235240c")
	if err != nil {
		t.Fatalf("Error: %v\n", err)
	}
	if ip != "10.211.55.124" {
		t.Fatalf("Should have found 10.211.55.124, not %s!\n", ip)
	}
}

func TestXMLParseConfig(t *testing.T) {
	td, err := ioutil.TempDir("", "configpvs")
	if err != nil {
		t.Fatalf("Error creating temp file: %s", err)
	}
	defer os.Remove(td)

	config := []byte(`
<ExampleParallelsConfig>
  <SystemConfig>
    <DiskSize>20</DiskSize>
  </SystemConfig>
</ExampleParallelsConfig>
`)
	ioutil.WriteFile(td+"/config.pvs", config, 0666)

	result, err := getConfigValueFromXpath(td, "//DiskSize")
	if err != nil {
		t.Fatalf("Error parsing XML: %s", err)
	}

	if result != "20" {
		t.Fatalf("Expected %q, got %q", "20", result)
	}
}
