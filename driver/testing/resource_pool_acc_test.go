package testing

import "testing"

func TestResourcePoolAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := NewTestDriver(t)
	p, err := d.FindResourcePool(TestHost, TestResourcePool)
	if err != nil {
		t.Fatalf("Cannot find the default resource pool '%v': %v", TestResourcePool, err)
	}
	CheckResourcePoolPath(t, p, TestResourcePool)
}
