package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func newRunner(steps []multistep.Step, config PackerConfig, ui packer.Ui) (multistep.Runner, multistep.DebugPauseFn) {
	switch config.PackerOnError {
	case "", "cleanup":
	case "abort":
		for i, step := range steps {
			steps[i] = abortStep{step, ui}
		}
	case "ask":
		for i, step := range steps {
			steps[i] = askStep{step, ui}
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
func NewRunner(steps []multistep.Step, config PackerConfig, ui packer.Ui) multistep.Runner {
	runner, _ := newRunner(steps, config, ui)
	return runner
}

// NewRunnerWithPauseFn returns a multistep.Runner that runs steps augmented
// with support for -debug and -on-error command line arguments.  With -debug it
// puts the multistep.DebugPauseFn that will pause execution between steps into
// the state under the key "pauseFn".
func NewRunnerWithPauseFn(steps []multistep.Step, config PackerConfig, ui packer.Ui, state multistep.StateBag) multistep.Runner {
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
	step multistep.Step
	ui   packer.Ui
}

func (s abortStep) InnerStepName() string {
	return typeName(s.step)
}

func (s abortStep) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	return s.step.Run(ctx, state)
}

func (s abortStep) Cleanup(state multistep.StateBag) {
	_, alreadyLogged := state.GetOk("abort_step_logged")
	err, ok := state.GetOk("error")
	if ok && !alreadyLogged {
		s.ui.Error(fmt.Sprintf("%s", err))
		state.Put("abort_step_logged", true)
	}
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		if !alreadyLogged {
			s.ui.Error("Interrupted, aborting...")
			state.Put("abort_step_logged", true)
		} else {
			s.ui.Error(fmt.Sprintf("aborted: skipping cleanup of step %q", typeName(s.step)))
		}
		return
	}
	if _, ok := state.GetOk(multistep.StateHalted); ok {
		if !alreadyLogged {
			s.ui.Error(fmt.Sprintf("Step %q failed, aborting...", typeName(s.step)))
			state.Put("abort_step_logged", true)
		} else {
			s.ui.Error(fmt.Sprintf("aborted: skipping cleanup of step %q", typeName(s.step)))
		}
		return
	}
	s.step.Cleanup(state)
}

type askStep struct {
	step multistep.Step
	ui   packer.Ui
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
			os.Exit(1)
		case askRetry:
			continue
		}
	}
}

func (s askStep) Cleanup(state multistep.StateBag) {
	s.step.Cleanup(state)
}

type askResponse int

const (
	askCleanup askResponse = iota
	askAbort
	askRetry
)

func ask(ui packer.Ui, name string, state multistep.StateBag) askResponse {
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

func askPrompt(ui packer.Ui) askResponse {
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
