package common

import "testing"

func TestVMnetNatConfIPFinder_Impl(t *testing.T) {
	var raw interface{}
	raw = &VMnetNatConfIPFinder{}
	if _, ok := raw.(HostIPFinder); !ok {
		t.Fatalf("VMnetNatConfIPFinder is not a host IP finder")
	}
}
