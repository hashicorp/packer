package common

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/packer/common/net"
	"github.com/hashicorp/packer/common/retry"
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
	session          *ssm.StartSessionOutput
}

// Run executes the Packer build step that creates a session tunnel.
func (s *StepCreateSSMTunnel) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.SSMAgentEnabled {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packer.Ui)
	if err := s.ConfigureLocalHostPort(ctx); err != nil {
		err := fmt.Errorf("error finding an available port to initiate a session tunnel: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	instance, ok := state.Get("instance").(*ec2.Instance)
	if !ok {
		err := fmt.Errorf("error encountered in obtaining target instance id for session tunnel")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	s.instanceId = aws.StringValue(instance.InstanceId)

	log.Printf("Starting PortForwarding session to instance %q on local port %q to remote port %q", s.instanceId, s.LocalPortNumber, s.RemotePortNumber)
	input := s.BuildTunnelInputForInstance(s.instanceId)
	ssmconn := ssm.New(s.AWSSession)
	var output *ssm.StartSessionOutput
	err := retry.Config{
		ShouldRetry: func(err error) bool { return isAWSErr(err, "TargetNotConnected", "") },
		RetryDelay:  (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) (err error) {
		output, err = ssmconn.StartSessionWithContext(ctx, &input)
		return err
	})

	if err != nil {
		err = fmt.Errorf("error encountered in starting session for instance %q: %s", s.instanceId, err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	driver := SSMDriver{
		Region:          s.Region,
		Session:         output,
		SessionParams:   input,
		SessionEndpoint: ssmconn.Endpoint,
	}

	if err := driver.StartSession(ctx); err != nil {
		err = fmt.Errorf("error encountered in establishing a tunnel %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("PortForwarding session tunnel to instance %q established!", s.instanceId))
	state.Put("sessionPort", s.LocalPortNumber)

	return multistep.ActionContinue
}

// Cleanup terminates an active session on AWS, which in turn terminates the associated tunnel process running on the local machine.
func (s *StepCreateSSMTunnel) Cleanup(state multistep.StateBag) {
	if s.session == nil {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ssmconn := ssm.New(s.AWSSession)
	_, err := ssmconn.TerminateSession(&ssm.TerminateSessionInput{SessionId: s.session.SessionId})
	if err != nil {
		msg := fmt.Sprintf("Error terminating SSM Session %q. Please terminate the session manually: %s",
			aws.StringValue(s.session.SessionId), err)
		ui.Error(msg)
	}

}

// ConfigureLocalHostPort finds an available port on the localhost that can be used for the remote tunnel.
// Defaults to using s.LocalPortNumber if it is set.
func (s *StepCreateSSMTunnel) ConfigureLocalHostPort(ctx context.Context) error {
	if s.LocalPortNumber != 0 {
		return nil
	}
	// Find an available TCP port for our HTTP server
	l, err := net.ListenRangeConfig{
		Min:     8000,
		Max:     9000,
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
