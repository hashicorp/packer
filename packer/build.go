package packer

import "log"

// A Build represents a single job within Packer that is responsible for
// building some machine image artifact. Builds are meant to be parallelized.
type Build interface {
	Name() string
	Prepare() error
	Run(Ui, Cache) (Artifact, error)
	Cancel()
}

// A build struct represents a single build job, the result of which should
// be a single machine image artifact. This artifact may be comprised of
// multiple files, of course, but it should be for only a single provider
// (such as VirtualBox, EC2, etc.).
type coreBuild struct {
	name          string
	builder       Builder
	builderConfig interface{}
	hooks         map[string][]Hook
	provisioners  []coreBuildProvisioner

	prepareCalled bool
}

// Keeps track of the provisioner and the configuration of the provisioner
// within the build.
type coreBuildProvisioner struct {
	provisioner Provisioner
	config      []interface{}
}

// Returns the name of the build.
func (b *coreBuild) Name() string {
	return b.name
}

// Prepare prepares the build by doing some initialization for the builder
// and any hooks. This _must_ be called prior to Run.
func (b *coreBuild) Prepare() (err error) {
	// TODO: lock
	b.prepareCalled = true

	// Prepare the builder
	err = b.builder.Prepare(b.builderConfig)
	if err != nil {
		log.Printf("Build '%s' prepare failure: %s\n", b.name, err)
		return
	}

	// Prepare the provisioners
	for _, coreProv := range b.provisioners {
		if err = coreProv.provisioner.Prepare(coreProv.config...); err != nil {
			return
		}
	}

	return
}

// Runs the actual build. Prepare must be called prior to running this.
func (b *coreBuild) Run(ui Ui, cache Cache) (Artifact, error) {
	if !b.prepareCalled {
		panic("Prepare must be called first")
	}

	// Copy the hooks
	hooks := make(map[string][]Hook)
	for hookName, hookList := range b.hooks {
		hooks[hookName] = make([]Hook, len(hookList))
		copy(hooks[hookName], hookList)
	}

	// Add a hook for the provisioners if we have provisioners
	if len(b.provisioners) > 0 {
		provisioners := make([]Provisioner, len(b.provisioners))
		for i, p := range b.provisioners {
			provisioners[i] = p.provisioner
		}

		if _, ok := hooks[HookProvision]; !ok {
			hooks[HookProvision] = make([]Hook, 0, 1)
		}

		hooks[HookProvision] = append(hooks[HookProvision], &ProvisionHook{provisioners})
	}

	hook := &DispatchHook{hooks}
	return b.builder.Run(ui, hook, cache)
}

// Cancels the build if it is running.
func (b *coreBuild) Cancel() {
	b.builder.Cancel()
}
