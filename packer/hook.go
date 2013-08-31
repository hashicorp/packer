package packer

import (
	"sync"
)

// This is the hook that should be fired for provisioners to run.
const HookProvision = "packer_provision"

// A Hook is used to hook into an arbitrarily named location in a build,
// allowing custom behavior to run at certain points along a build.
//
// Run is called when the hook is called, with the name of the hook and
// arbitrary data associated with it. To know what format the data is in,
// you must reference the documentation for the specific hook you're interested
// in. In addition to that, the Hook is given access to a UI so that it can
// output things to the user.
//
// Cancel is called when the hook needs to be cancelled. This will usually
// be called when Run is still in progress so the mechanism that handles this
// must be race-free. Cancel should attempt to cancel the hook in the
// quickest, safest way possible.
type Hook interface {
	Run(string, Ui, Communicator, interface{}) error
	Cancel()
}

// A Hook implementation that dispatches based on an internal mapping.
type DispatchHook struct {
	Mapping map[string][]Hook

	l           sync.Mutex
	cancelled   bool
	runningHook Hook
}

// Runs the hook with the given name by dispatching it to the proper
// hooks if a mapping exists. If a mapping doesn't exist, then nothing
// happens.
func (h *DispatchHook) Run(name string, ui Ui, comm Communicator, data interface{}) error {
	h.l.Lock()
	h.cancelled = false
	h.l.Unlock()

	// Make sure when we exit that we reset the running hook.
	defer func() {
		h.l.Lock()
		defer h.l.Unlock()
		h.runningHook = nil
	}()

	hooks, ok := h.Mapping[name]
	if !ok {
		// No hooks for that name. No problem.
		return nil
	}

	for _, hook := range hooks {
		h.l.Lock()
		if h.cancelled {
			h.l.Unlock()
			return nil
		}

		h.runningHook = hook
		h.l.Unlock()

		if err := hook.Run(name, ui, comm, data); err != nil {
			return err
		}
	}

	return nil
}

// Cancels all the hooks that are currently in-flight, if any. This will
// block until the hooks are all cancelled.
func (h *DispatchHook) Cancel() {
	h.l.Lock()
	defer h.l.Unlock()

	if h.runningHook != nil {
		h.runningHook.Cancel()
	}

	h.cancelled = true
}
