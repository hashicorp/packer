package arm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCertificateInKeyVault struct {
	config *Config
	client *AzureClient
	say    func(message string)
	error  func(e error)
}

func NewStepCertificateInKeyVault(cli *AzureClient, ui packer.Ui, config *Config) *StepCertificateInKeyVault {
	var step = &StepCertificateInKeyVault{
		client: cli,
		config: config,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	return step
}

func (s *StepCertificateInKeyVault) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Setting the certificate in the KeyVault...")

	var keyVaultName = state.Get(constants.ArmKeyVaultName).(string)
	// err := s.client.CreateKey(keyVaultName, DefaultSecretName)
	// if err != nil {
	// 	s.error(fmt.Errorf("Error setting winrm cert in custom keyvault: %s", err))
	// 	return multistep.ActionHalt
	// }

	err := s.client.SetSecret(keyVaultName, DefaultSecretName, s.config.winrmCertificate)
	if err != nil {
		s.error(fmt.Errorf("Error setting winrm cert in custom keyvault: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepCertificateInKeyVault) Cleanup(multistep.StateBag) {
}
