package vmware

import "testing"

func TestIfconfigIPFinder_Impl(t *testing.T) {
	var raw interface{}
	raw = &IfconfigIPFinder{}
	if _, ok := raw.(HostIPFinder); !ok {
		t.Fatalf("IfconfigIPFinder is not a host IP finder")
	}
}
