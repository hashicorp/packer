package common

import (
	"context"
	"encoding/json"
	"fmt"
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
	AWSSession      *session.Session
	DstPort         int
	SSMAgentEnabled bool
	instanceId      string
	ssmSession      *ssm.StartSessionOutput
}

func (s *StepCreateSSMTunnel) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.SSMAgentEnabled {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packer.Ui)
	// Find an available TCP port for our HTTP server
	l, err := net.ListenRangeConfig{
		Min:     8000,
		Max:     9000,
		Addr:    "0.0.0.0",
		Network: "tcp",
	}.Listen(ctx)
	if err != nil {
		err := fmt.Errorf("error finding an available port to initiate a session tunnel: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	dst, src := strconv.Itoa(s.DstPort), strconv.Itoa(l.Port)
	params := map[string][]*string{
		"portNumber":      []*string{aws.String(dst)},
		"localPortNumber": []*string{aws.String(src)},
	}
	l.Close()

	instance, ok := state.Get("instance").(*ec2.Instance)
	if !ok {
		err := fmt.Errorf("error encountered in obtaining target instance id for SSM tunnel")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	s.instanceId = aws.StringValue(instance.InstanceId)
	ssmconn := ssm.New(s.AWSSession)
	input := ssm.StartSessionInput{
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters:   params,
		Target:       aws.String(s.instanceId),
	}

	ui.Message(fmt.Sprintf("Starting PortForwarding session to instance %q on local port %q to remote port %q", s.instanceId, src, dst))
	var output *ssm.StartSessionOutput
	err = retry.Config{
		ShouldRetry: func(err error) bool { return isAWSErr(err, "TargetNotConnected", "") },
		RetryDelay:  (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		output, err = ssmconn.StartSessionWithContext(ctx, &input)
		return err
	})

	if err != nil {
		err = fmt.Errorf("error encountered in starting session for instance %q: %s", s.instanceId, err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	s.ssmSession = output

	// AWS session-manager-plugin requires a valid session be passed in JSON
	sessionDetails, err := json.Marshal(s.ssmSession)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error encountered in reading session details", err)
		return multistep.ActionHalt
	}

	sessionParameters, err := json.Marshal(input)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error encountered in reading session parameter details", err)
		return multistep.ActionHalt
	}

	driver := SSMDriver{Ui: ui}
	// sessionDetails, region, "StartSession", profile, paramJson, endpoint
	region := aws.StringValue(s.AWSSession.Config.Region)
	// how to best get Profile name
	if err := driver.StartSession(string(sessionDetails), region, "default", string(sessionParameters), ssmconn.Endpoint); err != nil {
		err = fmt.Errorf("error encountered in establishing a tunnel with the session-manager-plugin: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("PortForwarding session to instance %q established!", s.instanceId))
	state.Put("sessionPort", l.Port)

	return multistep.ActionContinue
}

func (s *StepCreateSSMTunnel) Cleanup(state multistep.StateBag) {
	if s.ssmSession == nil {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ssmconn := ssm.New(s.AWSSession)
	_, err := ssmconn.TerminateSession(&ssm.TerminateSessionInput{SessionId: s.ssmSession.SessionId})
	if err != nil {
		msg := fmt.Sprintf("Error terminating SSM Session %q. Please terminate the session manually: %s",
			aws.StringValue(s.ssmSession.SessionId), err)
		ui.Error(msg)
	}

}
