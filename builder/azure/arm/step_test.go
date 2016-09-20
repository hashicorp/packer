// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common"
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"testing"
)

func TestProcessStepResultShouldContinueForNonErrors(t *testing.T) {
	stateBag := new(multistep.BasicStateBag)

	code := processStepResult(nil, func(error) { t.Fatal("Should not be called!") }, stateBag)
	if _, ok := stateBag.GetOk(constants.Error); ok {
		t.Errorf("Error was nil, but was still in the state bag.")
	}

	if code != multistep.ActionContinue {
		t.Errorf("Expected ActionContinue(%d), but got=%d", multistep.ActionContinue, code)
	}
}

func TestProcessStepResultShouldHaltOnError(t *testing.T) {
	stateBag := new(multistep.BasicStateBag)
	isSaidError := false

	code := processStepResult(fmt.Errorf("boom"), func(error) { isSaidError = true }, stateBag)
	if _, ok := stateBag.GetOk(constants.Error); !ok {
		t.Errorf("Error was non nil, but was not in the state bag.")
	}

	if !isSaidError {
		t.Errorf("Expected error to be said, but it was not.")
	}

	if code != multistep.ActionHalt {
		t.Errorf("Expected ActionHalt(%d), but got=%d", multistep.ActionHalt, code)
	}
}

func TestProcessStepResultShouldContinueOnSuccessfulTask(t *testing.T) {
	stateBag := new(multistep.BasicStateBag)
	result := common.InterruptibleTaskResult{
		IsCancelled: false,
		Err:         nil,
	}

	code := processInterruptibleResult(result, func(error) { t.Fatal("Should not be called!") }, stateBag)
	if _, ok := stateBag.GetOk(constants.Error); ok {
		t.Errorf("Error was nil, but was still in the state bag.")
	}

	if code != multistep.ActionContinue {
		t.Errorf("Expected ActionContinue(%d), but got=%d", multistep.ActionContinue, code)
	}
}

func TestProcessStepResultShouldHaltWhenTaskIsCancelled(t *testing.T) {
	stateBag := new(multistep.BasicStateBag)
	result := common.InterruptibleTaskResult{
		IsCancelled: true,
		Err:         nil,
	}

	code := processInterruptibleResult(result, func(error) { t.Fatal("Should not be called!") }, stateBag)
	if _, ok := stateBag.GetOk(constants.Error); ok {
		t.Errorf("Error was nil, but was still in the state bag.")
	}

	if code != multistep.ActionHalt {
		t.Errorf("Expected ActionHalt(%d), but got=%d", multistep.ActionHalt, code)
	}
}

func TestProcessStepResultShouldHaltOnTaskError(t *testing.T) {
	stateBag := new(multistep.BasicStateBag)
	isSaidError := false
	result := common.InterruptibleTaskResult{
		IsCancelled: false,
		Err:         fmt.Errorf("boom"),
	}

	code := processInterruptibleResult(result, func(error) { isSaidError = true }, stateBag)
	if _, ok := stateBag.GetOk(constants.Error); !ok {
		t.Errorf("Error was *not* nil, but was not in the state bag.")
	}

	if !isSaidError {
		t.Errorf("Expected error to be said, but it was not.")
	}

	if code != multistep.ActionHalt {
		t.Errorf("Expected ActionHalt(%d), but got=%d", multistep.ActionHalt, code)
	}
}
