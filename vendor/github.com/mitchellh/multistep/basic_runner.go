package multistep

import (
	"sync"
	"sync/atomic"
)

type runState int32

const (
	stateIdle runState = iota
	stateRunning
	stateCancelling
)

// BasicRunner is a Runner that just runs the given slice of steps.
type BasicRunner struct {
	// Steps is a slice of steps to run. Once set, this should _not_ be
	// modified.
	Steps []Step

	cancelCh chan struct{}
	doneCh   chan struct{}
	state    runState
	l        sync.Mutex
}

func (b *BasicRunner) Run(state StateBag) {
	b.l.Lock()
	if b.state != stateIdle {
		panic("already running")
	}

	cancelCh := make(chan struct{})
	doneCh := make(chan struct{})
	b.cancelCh = cancelCh
	b.doneCh = doneCh
	b.state = stateRunning
	b.l.Unlock()

	defer func() {
		b.l.Lock()
		b.cancelCh = nil
		b.doneCh = nil
		b.state = stateIdle
		close(doneCh)
		b.l.Unlock()
	}()

	// This goroutine listens for cancels and puts the StateCancelled key
	// as quickly as possible into the state bag to mark it.
	go func() {
		select {
		case <-cancelCh:
			// Flag cancel and wait for finish
			state.Put(StateCancelled, true)
			<-doneCh
		case <-doneCh:
		}
	}()

	for _, step := range b.Steps {
		// We also check for cancellation here since we can't be sure
		// the goroutine that is running to set it actually ran.
		if runState(atomic.LoadInt32((*int32)(&b.state))) == stateCancelling {
			state.Put(StateCancelled, true)
			break
		}

		action := step.Run(state)
		defer step.Cleanup(state)

		if _, ok := state.GetOk(StateCancelled); ok {
			break
		}

		if action == ActionHalt {
			state.Put(StateHalted, true)
			break
		}
	}
}

func (b *BasicRunner) Cancel() {
	b.l.Lock()
	switch b.state {
	case stateIdle:
		// Not running, so Cancel is... done.
		b.l.Unlock()
		return
	case stateRunning:
		// Running, so mark that we cancelled and set the state
		close(b.cancelCh)
		b.state = stateCancelling
		fallthrough
	case stateCancelling:
		// Already cancelling, so just wait until we're done
		ch := b.doneCh
		b.l.Unlock()
		<-ch
	}
}
