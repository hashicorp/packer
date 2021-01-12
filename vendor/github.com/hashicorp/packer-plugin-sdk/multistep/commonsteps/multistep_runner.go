package commonsteps

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func newRunner(steps []multistep.Step, config common.PackerConfig, ui packersdk.Ui) (multistep.Runner, multistep.DebugPauseFn) {
	switch config.PackerOnError {
	case "", "cleanup":
	case "abort":
		for i, step := range steps {
			steps[i] = abortStep{
				step:        step,
				cleanupProv: false,
				ui:          ui,
			}
		}
	case "ask":
		for i, step := range steps {
			steps[i] = askStep{step, ui}
		}
	case "run-cleanup-provisioner":
		for i, step := range steps {
			steps[i] = abortStep{
				step:        step,
				cleanupProv: true,
				ui:          ui,
			}
		}
	}

	if config.PackerDebug {
		pauseFn := MultistepDebugFn(ui)
		return &multistep.DebugRunner{Steps: steps, PauseFn: pauseFn}, pauseFn
	} else {
		return &multistep.BasicRunner{Steps: steps}, nil
	}
}

// NewRunner returns a multistep.Runner that runs steps augmented with support
// for -debug and -on-error command line arguments.
func NewRunner(steps []multistep.Step, config common.PackerConfig, ui packersdk.Ui) multistep.Runner {
	runner, _ := newRunner(steps, config, ui)
	return runner
}

// NewRunnerWithPauseFn returns a multistep.Runner that runs steps augmented
// with support for -debug and -on-error command line arguments.  With -debug it
// puts the multistep.DebugPauseFn that will pause execution between steps into
// the state under the key "pauseFn".
func NewRunnerWithPauseFn(steps []multistep.Step, config common.PackerConfig, ui packersdk.Ui, state multistep.StateBag) multistep.Runner {
	runner, pauseFn := newRunner(steps, config, ui)
	if pauseFn != nil {
		state.Put("pauseFn", pauseFn)
	}
	return runner
}

func typeName(i interface{}) string {
	return reflect.Indirect(reflect.ValueOf(i)).Type().Name()
}

type abortStep struct {
	step        multistep.Step
	cleanupProv bool
	ui          packersdk.Ui
}

func (s abortStep) InnerStepName() string {
	return typeName(s.step)
}

func (s abortStep) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	return s.step.Run(ctx, state)
}

func (s abortStep) Cleanup(state multistep.StateBag) {
	if s.InnerStepName() == typeName(StepProvision{}) && s.cleanupProv {
		s.step.Cleanup(state)
		return
	}

	shouldCleanup := handleAbortsAndInterupts(state, s.ui, typeName(s.step))
	if !shouldCleanup {
		return
	}
	s.step.Cleanup(state)
}

type askStep struct {
	step multistep.Step
	ui   packersdk.Ui
}

func (s askStep) InnerStepName() string {
	return typeName(s.step)
}

func (s askStep) Run(ctx context.Context, state multistep.StateBag) (action multistep.StepAction) {
	for {
		action = s.step.Run(ctx, state)

		if action != multistep.ActionHalt {
			return
		}

		err, ok := state.GetOk("error")
		if ok {
			s.ui.Error(fmt.Sprintf("%s", err))
		}

		switch ask(s.ui, typeName(s.step), state) {
		case askCleanup:
			return
		case askAbort:
			state.Put("aborted", true)
			return
		case askRetry:
			continue
		}
	}
}

func (s askStep) Cleanup(state multistep.StateBag) {
	if _, ok := state.GetOk("aborted"); ok {
		shouldCleanup := handleAbortsAndInterupts(state, s.ui, typeName(s.step))
		if !shouldCleanup {
			return
		}
	}
	s.step.Cleanup(state)
}

type askResponse int

const (
	askCleanup askResponse = iota
	askAbort
	askRetry
)

func ask(ui packersdk.Ui, name string, state multistep.StateBag) askResponse {
	ui.Say(fmt.Sprintf("Step %q failed", name))

	result := make(chan askResponse)
	go func() {
		result <- askPrompt(ui)
	}()

	for {
		select {
		case response := <-result:
			return response
		case <-time.After(100 * time.Millisecond):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				return askCleanup
			}
		}
	}
}

func askPrompt(ui packersdk.Ui) askResponse {
	for {
		line, err := ui.Ask("[c] Clean up and exit, [a] abort without cleanup, or [r] retry step (build may fail even if retry succeeds)?")
		if err != nil {
			log.Printf("Error asking for input: %s", err)
		}

		input := strings.ToLower(line) + "c"
		switch input[0] {
		case 'c':
			return askCleanup
		case 'a':
			return askAbort
		case 'r':
			return askRetry
		}
		ui.Say(fmt.Sprintf("Incorrect input: %#v", line))
	}
}

func handleAbortsAndInterupts(state multistep.StateBag, ui packersdk.Ui, stepName string) bool {
	// if returns false, don't run cleanup. If true, do run cleanup.
	_, alreadyLogged := state.GetOk("abort_step_logged")

	err, ok := state.GetOk("error")
	if ok && !alreadyLogged {
		ui.Error(fmt.Sprintf("%s", err))
		state.Put("abort_step_logged", true)
	}
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		if !alreadyLogged {
			ui.Error("Interrupted, aborting...")
			state.Put("abort_step_logged", true)
		} else {
			ui.Error(fmt.Sprintf("aborted: skipping cleanup of step %q", stepName))
		}
		return false
	}
	if _, ok := state.GetOk(multistep.StateHalted); ok {
		if !alreadyLogged {
			ui.Error(fmt.Sprintf("Step %q failed, aborting...", stepName))
			state.Put("abort_step_logged", true)
		} else {
			ui.Error(fmt.Sprintf("aborted: skipping cleanup of step %q", stepName))
		}
		return false
	}
	return true
}
