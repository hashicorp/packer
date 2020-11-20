package vagrant

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

type StepAddBox struct {
	BoxVersion   string
	CACert       string
	CAPath       string
	DownloadCert string
	Clean        bool
	Force        bool
	Insecure     bool
	Provider     string
	SourceBox    string
	BoxName      string
	GlobalID     string
	SkipAdd      bool
}

func (s *StepAddBox) generateAddArgs() []string {
	addArgs := []string{}

	if strings.HasSuffix(s.SourceBox, ".box") {
		addArgs = append(addArgs, s.BoxName)
	}

	addArgs = append(addArgs, s.SourceBox)

	if s.BoxVersion != "" {
		addArgs = append(addArgs, "--box-version", s.BoxVersion)
	}

	if s.CACert != "" {
		addArgs = append(addArgs, "--cacert", s.CACert)
	}

	if s.CAPath != "" {
		addArgs = append(addArgs, "--capath", s.CAPath)
	}

	if s.DownloadCert != "" {
		addArgs = append(addArgs, "--cert", s.DownloadCert)
	}

	if s.Clean {
		addArgs = append(addArgs, "--clean")
	}

	if s.Force {
		addArgs = append(addArgs, "--force")
	}

	if s.Insecure {
		addArgs = append(addArgs, "--insecure")
	}

	if s.Provider != "" {
		addArgs = append(addArgs, "--provider", s.Provider)
	}

	return addArgs
}

func (s *StepAddBox) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(VagrantDriver)
	ui := state.Get("ui").(packer.Ui)

	if s.SkipAdd {
		ui.Say("skip_add was set so we assume the box is already in Vagrant...")
		return multistep.ActionContinue
	}

	if s.GlobalID != "" {
		ui.Say("Using a global-id; skipping Vagrant add command...")
		return multistep.ActionContinue
	}

	ui.Say("Adding box using vagrant box add ...")
	ui.Message("(this can take some time if we need to download the box)")
	addArgs := s.generateAddArgs()

	log.Printf("[vagrant] Calling box add with following args %s", strings.Join(addArgs, " "))
	// Call vagrant using prepared arguments
	err := driver.Add(addArgs)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepAddBox) Cleanup(state multistep.StateBag) {
}
