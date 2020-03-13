package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateSSMTunnel struct {
	CommConfig *communicator.Config
	AWSSession *session.Session
	InstanceID string
	DstPort    string
	SrcPort    string

	ssmSession *ssm.StartSessionOutput
	tunnel     net.Listener
}

func (s *StepCreateSSMTunnel) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	/*
		p, _ := strconv.Atoi(s.SrcPort)
		//TODO dynamically setup local port
		// Find an available TCP port for our HTTP server
		l, err := packernet.ListenRangeConfig{
			Min:     p,
			Max:     p,
			Addr:    "0.0.0.0",
			Network: "tcp",
		}.Listen(ctx)

		if err != nil {
			err := fmt.Errorf("Error finding port: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	*/
	params := map[string][]*string{
		"portNumber":      []*string{aws.String(s.DstPort)},
		"localPortNumber": []*string{aws.String(strconv.Itoa(8081))},
	}
	instance, ok := state.Get("instance").(*ec2.Instance)
	if !ok {
		err := fmt.Errorf("error encountered in obtaining target instance id for SSM tunnel")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	s.InstanceID = aws.StringValue(instance.InstanceId)

	ssmconn := ssm.New(s.AWSSession)
	input := ssm.StartSessionInput{
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters:   params,
		Target:       aws.String(s.InstanceID),
	}
	var output *ssm.StartSessionOutput
	var err error
	err = retry.Config{
		Tries:       11,
		ShouldRetry: func(err error) bool { return isAWSErr(err, "TargetNotConnected", "") },
		RetryDelay:  (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		output, err = ssmconn.StartSessionWithContext(ctx, &input)
		return err
	})

	if err != nil {
		err = fmt.Errorf("error encountered in starting session: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	s.ssmSession = output

	sessJson, err := json.Marshal(s.ssmSession)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	paramsJson, err := json.Marshal(input)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	driver := SSMDriver{Ui: ui}
	// sessJson, region, "StartSession", profile, paramJson, endpoint
	if err := driver.StartSession(string(sessJson), "us-east-1", "packer", string(paramsJson), ssmconn.Endpoint); err != nil {
		err = fmt.Errorf("error encountered in creating a connection to the SSM agent: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

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
