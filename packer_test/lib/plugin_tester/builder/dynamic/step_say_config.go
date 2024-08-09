// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This is a definition of a builder step and should implement multistep.Step
type StepSayConfig struct {
	cfg Config
}

// Run should execute the purpose of this step
func (s *StepSayConfig) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Dynamic builder invoked!")
	for _, nf := range s.cfg.Nesteds {
		ui.Say(fmt.Sprintf("Nested first: %s", nf.Name))
		for _, ns := range nf.Nesteds {
			ui.Say(fmt.Sprintf("Nested second: %s.%s", nf.Name, ns.Name))
		}
	}

	// Determines that should continue to the next step
	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepSayConfig) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}
