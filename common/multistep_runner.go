package common

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
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

func (s abortStep) Run(state multistep.StateBag) multistep.StepAction {
	return s.step.Run(state)
}

func (s abortStep) Cleanup(state multistep.StateBag) {
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		s.ui.Error("Interrupted, aborting...")
		os.Exit(1)
	}
	if _, ok := state.GetOk(multistep.StateHalted); ok {
		s.ui.Error(fmt.Sprintf("Step %q failed, aborting...", typeName(s.step)))
		os.Exit(1)
	}
	s.step.Cleanup(state)
}

type askStep struct {
	step multistep.Step
	ui   packer.Ui
}

func (s askStep) Run(state multistep.StateBag) (action multistep.StepAction) {
	for {
		action = s.step.Run(state)

		if action != multistep.ActionHalt {
			return
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
