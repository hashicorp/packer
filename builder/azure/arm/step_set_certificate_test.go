package arm

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepSetCertificateShouldPassIfGetPasses(t *testing.T) {
	var testSubject = &StepSetCertificate{
		config: new(Config),
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := createTestStateBagStepSetCertificate()

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepSetCertificateShouldTakeStepArgumentsFromStateBag(t *testing.T) {
	config := new(Config)
	var testSubject = &StepSetCertificate{
		config: config,
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := createTestStateBagStepSetCertificate()
	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if config.tmpWinRMCertificateUrl != stateBag.Get(constants.ArmCertificateUrl) {
		t.Fatalf("Expected config.tmpWinRMCertificateUrl to be %s, but got %s'", stateBag.Get(constants.ArmCertificateUrl), config.tmpWinRMCertificateUrl)
	}
}

func createTestStateBagStepSetCertificate() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)
	stateBag.Put(constants.ArmCertificateUrl, "Unit Test: Certificate URL")
	return stateBag
}
