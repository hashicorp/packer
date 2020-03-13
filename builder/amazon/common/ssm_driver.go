package common

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type SSMDriver struct {
	Ui  packer.Ui
	Ctx *interpolate.Context

	l sync.Mutex
}

// sessJson, region, "StartSession", profile, paramJson, endpoint
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

	// Remove log statement
	log.Printf("Attempting to start session with the following args: %v", args)
	cmd := exec.Command("session-manager-plugin", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		err = fmt.Errorf("Error committing container: %s\nStderr: %s", err, stderr.String())
		s.Ui.Error(err.Error())
		return err
	}
	log.Println(stdout.String())
	log.Println(stderr.String())

	return nil
}
