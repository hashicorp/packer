package common

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestStepCreateSSMTunnel_BuildTunnelInputForInstance(t *testing.T) {
	step := StepCreateSSMTunnel{
		Region:           "region",
		LocalPortNumber:  8001,
		RemotePortNumber: 22,
		SSMAgentEnabled:  true,
	}

	input := step.BuildTunnelInputForInstance("i-something")

	target := aws.StringValue(input.Target)
	if target != "i-something" {
		t.Errorf("input should contain instance id as target but it got %q", target)
	}

	params := map[string][]*string{
		"portNumber":      []*string{aws.String("22")},
		"localPortNumber": []*string{aws.String("8001")},
	}
	if !reflect.DeepEqual(input.Parameters, params) {
		t.Errorf("input should contain the expected port parameters but it got %v", input.Parameters)
	}

}

func TestStepCreateSSMTunnel_ConfigureLocalHostPort(t *testing.T) {
	tt := []struct {
		Name      string
		Step      StepCreateSSMTunnel
		PortCheck func(int) bool
	}{
		{"WithLocalPortNumber", StepCreateSSMTunnel{LocalPortNumber: 9001}, func(port int) bool { return port == 9001 }},
		{"WithNoLocalPortNumber", StepCreateSSMTunnel{}, func(port int) bool { return port >= 8000 && port <= 9000 }},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			step := tc.Step
			if err := step.ConfigureLocalHostPort(context.TODO()); err != nil {
				t.Errorf("failed to configure a port on localhost")
			}

			if !tc.PortCheck(step.LocalPortNumber) {
				t.Errorf("failed to configure a port on localhost")
			}
		})
	}

}
