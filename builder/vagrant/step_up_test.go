package vagrant

import (
	"strings"
	"testing"
)

func TestPrepUpArgs(t *testing.T) {
	type testArgs struct {
		Step     StepUp
		Expected []string
	}
	tests := []testArgs{
		{
			Step: StepUp{
				GlobalID: "foo",
				Provider: "bar",
			},
			Expected: []string{"foo", "--provider=bar"},
		},
		{
			Step:     StepUp{},
			Expected: []string{"source"},
		},
		{
			Step: StepUp{
				Provider: "pro",
			},
			Expected: []string{"source", "--provider=pro"},
		},
	}
	for _, test := range tests {
		args := test.Step.generateArgs()
		for i, val := range test.Expected {
			if strings.Compare(args[i], val) != 0 {
				t.Fatalf("expected %#v but received %#v", test.Expected, args)
			}
		}
	}
}
