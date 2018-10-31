package driver

import "testing"

func TestResourcePoolAcc(t *testing.T) {
	d := newTestDriver(t)
	p, err := d.FindResourcePool("", "esxi-1.vsphere65.test", "pool1/pool2")
	if err != nil {
		t.Fatalf("Cannot find the default resource pool '%v': %v", "pool1/pool2", err)
	}

	path, err := p.Path()
	if err != nil {
		t.Fatalf("Cannot read resource pool name: %v", err)
	}
	if path != "pool1/pool2" {
		t.Errorf("Wrong folder. expected: 'pool1/pool2', got: '%v'", path)
	}
}
