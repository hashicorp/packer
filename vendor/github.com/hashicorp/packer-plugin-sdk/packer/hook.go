package packer

import (
	"context"
)

// This is the hook that should be fired for provisioners to run.
const HookProvision = "packer_provision"
const HookCleanupProvision = "packer_cleanup_provision"

// A Hook is used to hook into an arbitrarily named location in a build,
// allowing custom behavior to run at certain points along a build.
//
// Run is called when the hook is called, with the name of the hook and
// arbitrary data associated with it. To know what format the data is in,
// you must reference the documentation for the specific hook you're interested
// in. In addition to that, the Hook is given access to a UI so that it can
// output things to the user.
//
// The first context argument controls cancellation, the context will usually
// be called when Run is still in progress so the mechanism that handles this
// must be race-free. Cancel should attempt to cancel the hook in the quickest,
// safest way possible.
type Hook interface {
	Run(context.Context, string, Ui, Communicator, interface{}) error
}

// A Hook implementation that dispatches based on an internal mapping.
type DispatchHook struct {
	Mapping map[string][]Hook
}

// Runs the hook with the given name by dispatching it to the proper
// hooks if a mapping exists. If a mapping doesn't exist, then nothing
// happens.
func (h *DispatchHook) Run(ctx context.Context, name string, ui Ui, comm Communicator, data interface{}) error {
	hooks, ok := h.Mapping[name]
	if !ok {
		// No hooks for that name. No problem.
		return nil
	}

	for _, hook := range hooks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := hook.Run(ctx, name, ui, comm, data); err != nil {
			return err
		}
	}

	return nil
}
