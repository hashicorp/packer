package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/aws/aws-sdk-go/service/ssm"
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
	// Provided for testing purposes; if not specified it defaults to sessionManagerPluginName
	PluginName string
}

// StartSession starts an interactive Systems Manager session with a remote instance via the AWS session-manager-plugin
func (sd *SSMDriver) StartSession(ctx context.Context) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if sd.PluginName == "" {
		sd.PluginName = sessionManagerPluginName
	}

	args, err := sd.Args()
	if err != nil {
		err = fmt.Errorf("error encountered validating session details: %s", err)
		return err
	}

	cmd := exec.CommandContext(ctx, sd.PluginName, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		err = fmt.Errorf("error encountered when calling %s: %s\nStderr: %s", sd.PluginName, err, stderr.String())
		return err
	}
	// TODO capture logging for testing
	log.Println(stdout.String())

	return nil
}
func (sd *SSMDriver) Args() ([]string, error) {
	if sd.Session == nil {
		return nil, fmt.Errorf("an active Amazon SSM Session is required before trying to open a session tunnel")
	}

	// AWS session-manager-plugin requires a valid session be passed in JSON.
	sessionDetails, err := json.Marshal(sd.Session)
	if err != nil {
		return nil, fmt.Errorf("error encountered in reading session details %s", err)
	}

	// AWS session-manager-plugin requires the parameters used in the session to be passed in JSON as well.
	sessionParameters, err := json.Marshal(sd.SessionParams)
	if err != nil {
		return nil, fmt.Errorf("error encountered in reading session parameter details %s", err)
	}

	args := []string{
		string(sessionDetails),
		sd.Region,
		sessionCommand,
		sd.ProfileName,
		string(sessionParameters),
		sd.SessionEndpoint,
	}

	return args, nil
}
