package arm

import (
	"context"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepSetCertificate struct {
	config *Config
	say    func(message string)
	error  func(e error)
}

func NewStepSetCertificate(config *Config, ui packersdk.Ui) *StepSetCertificate {
	var step = &StepSetCertificate{
		config: config,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	return step
}

func (s *StepSetCertificate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Setting the certificate's URL ...")

	var winRMCertificateUrl = state.Get(constants.ArmCertificateUrl).(string)
	s.config.tmpWinRMCertificateUrl = winRMCertificateUrl

	return multistep.ActionContinue
}

func (*StepSetCertificate) Cleanup(multistep.StateBag) {
}
