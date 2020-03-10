package common

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
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
}

func (s *StepCreateSSMTunnel) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	params := map[string][]*string{
		"portNumber":      []*string{aws.String(s.DstPort)},
		"localPortNumber": []*string{aws.String(s.SrcPort)},
	}
	ssmconn := ssm.New(s.AWSSession)
	input := ssm.StartSessionInput{
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters:   params,
		Target:       aws.String(s.InstanceID),
	}

	output, err := ssmconn.StartSession(&input)
	if err != nil {
		err = fmt.Errorf("error encountered in creating a connection to the SSM agent: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	s.ssmSession = output

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
