// multistep is a library for bulding up complex actions using individual,
// discrete steps.
package multistep

// A StepAction determines the next step to take regarding multi-step actions.
type StepAction uint

const (
	ActionContinue StepAction = iota
	ActionHalt
)

// This is the key set in the state bag when using the basic runner to
// signal that the step sequence was cancelled.
const StateCancelled = "cancelled"

// This is the key set in the state bag when a step halted the sequence.
const StateHalted = "halted"

// Step is a single step that is part of a potentially large sequence
// of other steps, responsible for performing some specific action.
type Step interface {
	// Run is called to perform the action. The parameter is a "state bag"
	// of untyped things. Please be very careful about type-checking the
	// items in this bag.
	//
	// The return value determines whether multi-step sequences continue
	// or should halt.
	Run(StateBag) StepAction

	// Cleanup is called in reverse order of the steps that have run
	// and allow steps to clean up after themselves. Do not assume if this
	// ran that the entire multi-step sequence completed successfully. This
	// method can be ran in the face of errors and cancellations as well.
	//
	// The parameter is the same "state bag" as Run, and represents the
	// state at the latest possible time prior to calling Cleanup.
	Cleanup(StateBag)
}

// Runner is a thing that runs one or more steps.
type Runner interface {
	// Run runs the steps with the given initial state.
	Run(StateBag)

	// Cancel cancels a potentially running stack of steps.
	Cancel()
}
