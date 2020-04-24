package common

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/hashicorp/packer/packer"
)

const SessionManagerPluginName string = "session-manager-plugin"

type SSMDriver struct {
	Ui  packer.Ui
	Ctx context.Context
	// Provided for testing purposes; if not specified it defaults to SessionManagerPluginName
	PluginName string
}

// StartSession starts an interactive Systems Manager session with a remote instance via the AWS session-manager-plugin
func (s *SSMDriver) StartSession(sessionData, region, profile, params, endpoint string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := []string{
		sessionData,
		region,
		"StartSession",
		profile,
		params,
		endpoint,
	}

	if s.PluginName == "" {
		s.PluginName = SessionManagerPluginName
	}

	if _, err := exec.LookPath(s.PluginName); err != nil {
		return err
	}

	log.Printf("Attempting to start session with the following args: %v", args)
	cmd := exec.CommandContext(s.Ctx, s.PluginName, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		err = fmt.Errorf("error encountered when calling %s: %s\nStderr: %s", s.PluginName, err, stderr.String())
		s.Ui.Error(err.Error())
		return err
	}
	// TODO capture logging for testing
	log.Println(stdout.String())

	return nil
}
