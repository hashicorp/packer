package shell

import (
	"fmt"
	"testing"
)

func TestProvisioner_ValidExitCode(t *testing.T) {

	tests := []struct {
		exitCodes []int
		code      int
		wantErr   bool
	}{
		{nil, 0, false},
		{nil, 1, true},
		{[]int{2}, 2, false},
		{[]int{2}, 3, true},
	}
	for n := range tests {
		tt := tests[n]
		t.Run(fmt.Sprintf("%v - %v - %v", tt.exitCodes, tt.code, tt.wantErr), func(t *testing.T) {
			p := Provisioner{
				ValidExitCodes: tt.exitCodes,
			}
			if err := p.ValidExitCode(tt.code); (err != nil) != tt.wantErr {
				t.Errorf("Provisioner.ValidExitCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
