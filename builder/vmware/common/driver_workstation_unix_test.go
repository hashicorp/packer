// +build !windows

package common

import (
	"testing"
)

func TestWorkstationVersion_ws14(t *testing.T) {
	input := `VMware Workstation Information:
VMware Workstation 14.1.1 build-7528167 Release`
	if err := workstationTestVersion("10", input); err != nil {
		t.Fatal(err)
	}
}
