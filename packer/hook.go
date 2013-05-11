package packer

// A Hook is used to hook into an arbitrarily named location in a build,
// allowing custom behavior to run at certain points along a build.
//
// Run is called when the hook is called, with the name of the hook and
// arbitrary data associated with it. To know what format the data is in,
// you must reference the documentation for the specific hook you're interested
// in. In addition to that, the Hook is given access to a UI so that it can
// output things to the user.
type Hook interface {
	Run(string, interface{}, Ui)
}

// A Hook implementation that dispatches based on an internal mapping.
type DispatchHook struct {
	Mapping map[string][]Hook
}

// Runs the hook with the given name by dispatching it to the proper
// hooks if a mapping exists. If a mapping doesn't exist, then nothing
// happens.
func (h *DispatchHook) Run(name string, data interface{}, ui Ui) {
	hooks, ok := h.Mapping[name]
	if !ok {
		// No hooks for that name. No problem.
		return
	}

	for _, hook := range hooks {
		hook.Run(name, data, ui)
	}
}
