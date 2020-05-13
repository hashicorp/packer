package common

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/packer"
)

func TestStepCreateSSMTunnel_Run(t *testing.T) {
	mockSvc := MockSSMSvc{}
	config := SSMDriverConfig{
		SvcClient:   &mockSvc,
		SvcEndpoint: "example.com",
	}

	mockDriver := NewSSMDriver(config)
	mockDriver.pluginCmdFunc = MockPluginCmdFunc

	state := testState()
	state.Put("ui", &packer.NoopUi{})
	state.Put("instance", &ec2.Instance{InstanceId: aws.String("i-something")})

	step := StepCreateSSMTunnel{
		driver: mockDriver,
	}

	step.Run(context.Background(), state)

	err := state.Get("error")
	if err != nil {
		err = err.(error)
		t.Fatalf("the call to Run failed with an error when it should've executed: %v", err)
	}

	if mockSvc.StartSessionCalled {
		t.Errorf("StartSession should not be called when SSMAgentEnabled is false")
	}

	// Run when SSMAgentEnabled is true
	step.SSMAgentEnabled = true
	step.Run(context.Background(), state)

	err = state.Get("error")
	if err != nil {
		err = err.(error)
		t.Fatalf("the call to Run failed with an error when it should've executed: %v", err)
	}

	if !mockSvc.StartSessionCalled {
		t.Errorf("calling run with the correct inputs should call StartSession")
	}

	step.Cleanup(state)
	if !mockSvc.TerminateSessionCalled {
		t.Errorf("calling cleanup on a successful run should call TerminateSession")
	}
}

func TestStepCreateSSMTunnel_Cleanup(t *testing.T) {
	mockSvc := MockSSMSvc{}
	config := SSMDriverConfig{
		SvcClient:   &mockSvc,
		SvcEndpoint: "example.com",
	}

	mockDriver := NewSSMDriver(config)
	mockDriver.pluginCmdFunc = MockPluginCmdFunc

	step := StepCreateSSMTunnel{
		SSMAgentEnabled: true,
		driver:          mockDriver,
	}

	state := testState()
	state.Put("ui", &packer.NoopUi{})
	state.Put("instance", &ec2.Instance{InstanceId: aws.String("i-something")})

	step.Cleanup(state)

	if mockSvc.TerminateSessionCalled {
		t.Fatalf("calling cleanup on a non started session should not call TerminateSession")
	}

}

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
