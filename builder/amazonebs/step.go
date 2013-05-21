package amazonebs

// A StepAction determines the next step to take regarding multi-step actions.
type StepAction uint

const (
	StepContinue StepAction = iota
	StepHalt
)

// Step is a single step that is part of a potentially large sequence
// of other steps, responsible for performing some specific action.
type Step interface {
	// Run is called to perform the action. The parameter is a "state bag"
	// of untyped things. Please be very careful about type-checking the
	// items in this bag.
	//
	// The return value determines whether multi-step sequences continue
	// or should halt.
	Run(map[string]interface{}) StepAction

	// Cleanup is called in reverse order of the steps that have run
	// and allow steps to clean up after themselves.
	//
	// The parameter is the same "state bag" as Run.
	Cleanup(map[string]interface{})
}

// RunSteps runs a sequence of steps.
func RunSteps(state map[string]interface{}, steps []Step) {
	for _, step := range steps {
		action := step.Run(state)
		defer step.Cleanup(state)

		if action == StepHalt {
			break
		}
	}
}
