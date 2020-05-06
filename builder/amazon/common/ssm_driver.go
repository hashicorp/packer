package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/mitchellh/iochan"
)

const sessionManagerPluginName string = "session-manager-plugin"

//sessionCommand is the AWS-SDK equivalent to the command you would specify to `aws ssm ...`
const sessionCommand string = "StartSession"

type SSMDriver struct {
	Region          string
	ProfileName     string
	Session         *ssm.StartSessionOutput
	SessionParams   ssm.StartSessionInput
	SessionEndpoint string
	PluginName      string
}

// StartSession starts an interactive Systems Manager session with a remote instance via the AWS session-manager-plugin
func (d *SSMDriver) StartSession(ctx context.Context) error {
	if d.PluginName == "" {
		d.PluginName = sessionManagerPluginName
	}

	args, err := d.Args()
	if err != nil {
		err = fmt.Errorf("error encountered validating session details: %s", err)
		return err
	}

	cmd := exec.CommandContext(ctx, d.PluginName, args...)

	// Let's build up our logging
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Create the channels we'll use for data
	stdoutCh := iochan.DelimReader(stdout, '\n')
	stderrCh := iochan.DelimReader(stderr, '\n')

	// Loop and get all our output
	go func(ctx context.Context, prefix string) {
		for {
			select {
			case <-ctx.Done():
				return
			case output := <-stderrCh:
				if output != "" {
					log.Printf("[ERROR] %s: %s", prefix, output)
				}
			case output := <-stdoutCh:
				if output != "" {
					log.Printf("[DEBUG] %s: %s", prefix, output)
				}
			}
		}
	}(ctx, d.PluginName)

	log.Printf("[DEBUG %s] opening session tunnel to instance %q for session %q", d.PluginName, aws.StringValue(d.SessionParams.Target), aws.StringValue(d.Session.SessionId))
	if err := cmd.Start(); err != nil {
		err = fmt.Errorf("error encountered when calling %s: %s\n", d.PluginName, err)
		return err
	}

	return nil
}

func (d *SSMDriver) Args() ([]string, error) {
	if d.Session == nil {
		return nil, fmt.Errorf("an active Amazon SSM Session is required before trying to open a session tunnel")
	}

	// AWS session-manager-plugin requires a valid session be passed in JSON.
	sessionDetails, err := json.Marshal(d.Session)
	if err != nil {
		return nil, fmt.Errorf("error encountered in reading session details %s", err)
	}

	// AWS session-manager-plugin requires the parameters used in the session to be passed in JSON as well.
	sessionParameters, err := json.Marshal(d.SessionParams)
	if err != nil {
		return nil, fmt.Errorf("error encountered in reading session parameter details %s", err)
	}

	// Args must be in this order
	args := []string{
		string(sessionDetails),
		d.Region,
		sessionCommand,
		d.ProfileName,
		string(sessionParameters),
		d.SessionEndpoint,
	}

	return args, nil
}
