package multistep

import "golang.org/x/net/context"

// A step for testing that accumuluates data into a string slice in the
// the state bag. It always uses the "data" key in the state bag, and will
// initialize it.
type TestStepAcc struct {
	// The data inserted into the state bag.
	Data string

	// If true, it will halt at the step when it is run
	Halt bool
}

// A step that syncs by sending a channel and expecting a response.
type TestStepSync struct {
	Ch chan chan bool
}

// A step that sleeps forever
type TestStepWaitForever struct {
}

// A step that manually flips state to cancelling in run
type TestStepInjectCancel struct {
}

func (s TestStepAcc) Run(_ context.Context, state StateBag) StepAction {
	s.insertData(state, "data")

	if s.Halt {
		return ActionHalt
	}

	return ActionContinue
}

func (s TestStepAcc) Cleanup(state StateBag) {
	s.insertData(state, "cleanup")
}

func (s TestStepAcc) insertData(state StateBag, key string) {
	if _, ok := state.GetOk(key); !ok {
		state.Put(key, make([]string, 0, 5))
	}

	data := state.Get(key).([]string)
	data = append(data, s.Data)
	state.Put(key, data)
}

func (s TestStepSync) Run(context.Context, StateBag) StepAction {
	ch := make(chan bool)
	s.Ch <- ch
	<-ch

	return ActionContinue
}

func (s TestStepSync) Cleanup(StateBag) {}

func (s TestStepWaitForever) Run(context.Context, StateBag) StepAction {
	select {}
}

func (s TestStepWaitForever) Cleanup(StateBag) {}

func (s TestStepInjectCancel) Run(_ context.Context, state StateBag) StepAction {
	r := state.Get("runner").(*BasicRunner)
	r.state = stateCancelling
	return ActionContinue
}

func (s TestStepInjectCancel) Cleanup(StateBag) {}
