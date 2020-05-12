package common

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/packer/common/net"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateSSMTunnel struct {
	AWSSession       *session.Session
	Region           string
	LocalPortNumber  int
	RemotePortNumber int
	SSMAgentEnabled  bool
	instanceId       string
	driver           *SSMDriver
}

// Run executes the Packer build step that creates a session tunnel.
func (s *StepCreateSSMTunnel) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if !s.SSMAgentEnabled {
		return multistep.ActionContinue
	}

	// Configure local port number
	if err := s.ConfigureLocalHostPort(ctx); err != nil {
		err := fmt.Errorf("error finding an available port to initiate a session tunnel: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Get instance information
	instance, ok := state.Get("instance").(*ec2.Instance)
	if !ok {
		err := fmt.Errorf("error encountered in obtaining target instance id for session tunnel")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	s.instanceId = aws.StringValue(instance.InstanceId)

	if s.driver == nil {
		ssmconn := ssm.New(s.AWSSession)
		cfg := SSMDriverConfig{
			SvcClient:   ssmconn,
			Region:      s.Region,
			SvcEndpoint: ssmconn.Endpoint,
		}
		driver := SSMDriver{SSMDriverConfig: cfg}
		s.driver = &driver
	}

	input := s.BuildTunnelInputForInstance(s.instanceId)
	_, err := s.driver.StartSession(ctx, input)
	if err != nil {
		err = fmt.Errorf("error encountered in establishing a tunnel %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("PortForwarding session %q has been started", s.instanceId))
	state.Put("sessionPort", s.LocalPortNumber)
	return multistep.ActionContinue
}

// Cleanup terminates an active session on AWS, which in turn terminates the associated tunnel process running on the local machine.
func (s *StepCreateSSMTunnel) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	if !s.SSMAgentEnabled {
		return
	}

	if err := s.driver.StopSession(); err != nil {
		ui.Error(err.Error())
	}
}

// ConfigureLocalHostPort finds an available port on the localhost that can be used for the remote tunnel.
// Defaults to using s.LocalPortNumber if it is set.
func (s *StepCreateSSMTunnel) ConfigureLocalHostPort(ctx context.Context) error {
	minPortNumber, maxPortNumber := 8000, 9000

	if s.LocalPortNumber != 0 {
		minPortNumber = s.LocalPortNumber
		maxPortNumber = minPortNumber
	}

	// Find an available TCP port for our HTTP server
	l, err := net.ListenRangeConfig{
		Min:     minPortNumber,
		Max:     maxPortNumber,
		Addr:    "0.0.0.0",
		Network: "tcp",
	}.Listen(ctx)
	if err != nil {
		return err
	}

	s.LocalPortNumber = l.Port
	// Stop listening on selected port so that the AWS session-manager-plugin can use it.
	// The port is closed right before we start the session to avoid two Packer builds from getting the same port - fingers-crossed
	l.Close()

	return nil

}

func (s *StepCreateSSMTunnel) BuildTunnelInputForInstance(instance string) ssm.StartSessionInput {
	dst, src := strconv.Itoa(s.RemotePortNumber), strconv.Itoa(s.LocalPortNumber)
	params := map[string][]*string{
		"portNumber":      []*string{aws.String(dst)},
		"localPortNumber": []*string{aws.String(src)},
	}

	input := ssm.StartSessionInput{
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters:   params,
		Target:       aws.String(instance),
	}

	return input
}
