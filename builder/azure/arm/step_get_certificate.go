package arm

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepGetCertificate struct {
	client *AzureClient
	get    func(keyVaultName string, secretName string) (string, error)
	say    func(message string)
	error  func(e error)
	pause  func()
}

func NewStepGetCertificate(client *AzureClient, ui packersdk.Ui) *StepGetCertificate {
	var step = &StepGetCertificate{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
		pause:  func() { time.Sleep(30 * time.Second) },
	}

	step.get = step.getCertificateUrl
	return step
}

func (s *StepGetCertificate) getCertificateUrl(keyVaultName string, secretName string) (string, error) {
	secret, err := s.client.GetSecret(keyVaultName, secretName)
	if err != nil {
		s.say(s.client.LastError.Error())
		return "", err
	}

	return *secret.ID, err
}

func (s *StepGetCertificate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Getting the certificate's URL ...")

	var keyVaultName = state.Get(constants.ArmKeyVaultName).(string)

	s.say(fmt.Sprintf(" -> Key Vault Name        : '%s'", keyVaultName))
	s.say(fmt.Sprintf(" -> Key Vault Secret Name : '%s'", DefaultSecretName))

	var err error
	var url string
	for i := 0; i < 5; i++ {
		url, err = s.get(keyVaultName, DefaultSecretName)
		if err == nil {
			break
		}

		s.say(fmt.Sprintf(" ...failed to get certificate URL, retry(%d)", i))
		s.pause()
	}

	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	s.say(fmt.Sprintf(" -> Certificate URL       : '%s'", url))
	state.Put(constants.ArmCertificateUrl, url)

	return multistep.ActionContinue
}

func (*StepGetCertificate) Cleanup(multistep.StateBag) {
}
