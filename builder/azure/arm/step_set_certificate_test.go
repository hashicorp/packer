// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"testing"

	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
)

func TestStepSetCertificateShouldPassIfGetPasses(t *testing.T) {
	var testSubject = &StepSetCertificate{
		config: new(Config),
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := createTestStateBagStepSetCertificate()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepSetCertificateShouldTakeStepArgumentsFromStateBag(t *testing.T) {
	var testSubject = &StepSetCertificate{
		config: new(Config),
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := createTestStateBagStepSetCertificate()
	var result = testSubject.Run(stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	_, ok := stateBag.GetOk(constants.ArmTemplateParameters)
	if !ok {
		t.Fatalf("Expected the state bag to have a value for '%s', but it did not.", constants.ArmTemplateParameters)
	}
}

func createTestStateBagStepSetCertificate() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)
	stateBag.Put(constants.ArmCertificateUrl, "Unit Test: Certificate URL")
	return stateBag
}
