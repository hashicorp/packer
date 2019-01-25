package vagrant

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
}

func (s *StepAddBox) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(VagrantDriver)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	ui.Say("Adding box using vagrant box add..")
	addArgs := []string{}

	if strings.HasSuffix(s.SourceBox, ".box") {
		// The box isn't a namespace like you'd pull from vagrant cloud
		if s.BoxName == "" {
			s.BoxName = fmt.Sprintf("packer_%s", config.PackerBuildName)
		}

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
