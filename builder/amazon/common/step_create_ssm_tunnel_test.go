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
	tun := StepCreateSSMTunnel{}

	ctx := context.TODO()
	if err := tun.ConfigureLocalHostPort(ctx); err != nil {
		t.Errorf("failed to configure a port on localhost")
	}

	if tun.LocalPortNumber == 0 {
		t.Errorf("failed to configure a port on localhost")
	}

}
