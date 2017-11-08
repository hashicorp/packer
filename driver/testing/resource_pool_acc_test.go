package testing

import "testing"

func TestResourcePoolAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := NewTestDriver(t)
	p, err := d.FindResourcePool("esxi-1.vsphere55.test", "pool1/pool2")
	if err != nil {
		t.Fatalf("Cannot find the default resource pool '%v': %v", "pool1/pool2", err)
	}
	CheckResourcePoolPath(t, p, "pool1/pool2")
}
