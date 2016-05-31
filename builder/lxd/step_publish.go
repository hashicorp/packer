package lxd

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

type stepPublish struct{}

func (s *stepPublish) Run(state multistep.StateBag) multistep.StepAction {
	var stdout, stderr bytes.Buffer
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName

	args := []string{
		// If we use `lxc stop <container>`, an ephemeral container would die forever.
		// `lxc publish` has special logic to handle this case.
		"lxc", "publish", "--force", name, "--alias", config.OutputImage,
	}

	ui.Say("Publishing container...")
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("Sudo command error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	r := regexp.MustCompile("([0-9a-fA-F]+)$")
	fingerprint := r.FindAllStringSubmatch(stdoutString, -1)[0][0]

	ui.Say(fmt.Sprintf("Created image: %s", fingerprint))

	state.Put("imageFingerprint", fingerprint)

	return multistep.ActionContinue
}

func (s *stepPublish) Cleanup(state multistep.StateBag) {}
