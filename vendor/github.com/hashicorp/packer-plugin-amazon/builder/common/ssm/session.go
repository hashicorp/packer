package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/hashicorp/packer-plugin-amazon/builder/common/awserrors"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer-plugin-sdk/shell-local/localexec"
)

type Session struct {
	SvcClient             ssmiface.SSMAPI
	Region                string
	InstanceID            string
	LocalPort, RemotePort int
}

func (s Session) buildTunnelInput() *ssm.StartSessionInput {
	portNumber, localPortNumber := strconv.Itoa(s.RemotePort), strconv.Itoa(s.LocalPort)
	params := map[string][]*string{
		"portNumber":      []*string{aws.String(portNumber)},
		"localPortNumber": []*string{aws.String(localPortNumber)},
	}

	return &ssm.StartSessionInput{
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters:   params,
		Target:       aws.String(s.InstanceID),
	}
}

// getCommand return a valid ordered set of arguments to pass to the driver command.
func (s Session) getCommand(ctx context.Context) ([]string, string, error) {
	input := s.buildTunnelInput()

	var session *ssm.StartSessionOutput
	err := retry.Config{
		ShouldRetry: func(err error) bool { return awserrors.Matches(err, "TargetNotConnected", "") },
		RetryDelay:  (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) (err error) {
		session, err = s.SvcClient.StartSessionWithContext(ctx, input)
		return err
	})

	if err != nil {
		return nil, "", err
	}

	if session == nil {
		return nil, "", fmt.Errorf("an active Amazon SSM Session is required before trying to open a session tunnel")
	}

	// AWS session-manager-plugin requires a valid session be passed in JSON.
	sessionDetails, err := json.Marshal(session)
	if err != nil {
		return nil, *session.SessionId, fmt.Errorf("error encountered in reading session details %s", err)
	}

	// AWS session-manager-plugin requires the parameters used in the session to be passed in JSON as well.
	sessionParameters, err := json.Marshal(input)
	if err != nil {
		return nil, "", fmt.Errorf("error encountered in reading session parameter details %s", err)
	}

	// Args must be in this order
	args := []string{
		string(sessionDetails),
		s.Region,
		"StartSession",
		"", // ProfileName
		string(sessionParameters),
		*session.StreamUrl,
	}
	return args, *session.SessionId, nil
}

// Start an interactive Systems Manager session with a remote instance via the
// AWS session-manager-plugin. To terminate the session you must cancell the
// context. If you do not wish to terminate the session manually: calling
// StopSession on a instance of this driver will terminate the active session
// created from calling StartSession.
func (s Session) Start(ctx context.Context, ui packersdk.Ui) error {
	for ctx.Err() == nil {
		log.Printf("ssm: Starting PortForwarding session to instance %s", s.InstanceID)
		args, sessionID, err := s.getCommand(ctx)
		if sessionID != "" {
			defer func() {
				_, err := s.SvcClient.TerminateSession(&ssm.TerminateSessionInput{SessionId: aws.String(sessionID)})
				if err != nil {
					ui.Error(fmt.Sprintf("Error terminating SSM Session %q. Please terminate the session manually: %s", sessionID, err))
				}
			}()
		}
		if err != nil {
			return err
		}

		cmd := exec.CommandContext(ctx, "session-manager-plugin", args...)

		ui.Message(fmt.Sprintf("Starting portForwarding session %q.", sessionID))
		err = localexec.RunAndStream(cmd, ui, nil)
		if err != nil {
			ui.Error(err.Error())
		}
	}
	return nil
}
