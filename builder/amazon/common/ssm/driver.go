package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/builder/localexec"
	"github.com/hashicorp/packer/packer"
)

type Session struct {
	SvcClient ssmiface.SSMAPI
	Region    string
	Input     ssm.StartSessionInput
}

// Returns true if the error matches all these conditions:
//  * err is of type awserr.Error
//  * Error.Code() matches code
//  * Error.Message() contains message
func isAWSErr(err error, code string, message string) bool {
	if err, ok := err.(awserr.Error); ok {
		return err.Code() == code && strings.Contains(err.Message(), message)
	}
	return false
}

// getCommand return a valid ordered set of arguments to pass to the driver command.
func (s Session) getCommand(ctx context.Context) ([]string, string, error) {
	var session *ssm.StartSessionOutput
	err := retry.Config{
		ShouldRetry: func(err error) bool { return isAWSErr(err, "TargetNotConnected", "") },
		RetryDelay:  (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) (err error) {
		session, err = s.SvcClient.StartSessionWithContext(ctx, &s.Input)
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
	sessionParameters, err := json.Marshal(s.Input)
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
func (s Session) Start(ctx context.Context, ui packer.Ui) error {
	for ctx.Err() == nil {
		log.Printf("ssm: Starting PortForwarding session to instance %q", *s.Input.Target)
		args, sessionID, err := s.getCommand(ctx)
		if sessionID != "" {
			defer func() {
				_, err := s.SvcClient.TerminateSession(&ssm.TerminateSessionInput{SessionId: aws.String(sessionID)})
				if err != nil {
					err = fmt.Errorf("Error terminating SSM Session %q. Please terminate the session manually: %s", sessionID, err)
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
