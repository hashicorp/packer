package packer

import (
	"log"
	"sync"
)

// This is the key in configurations that is set to "true" when Packer
// debugging is enabled.
const DebugConfigKey = "packer_debug"

// A Build represents a single job within Packer that is responsible for
// building some machine image artifact. Builds are meant to be parallelized.
type Build interface {
	// Name is the name of the build. This is unique across a single template,
	// but not absolutely unique. This is meant more to describe to the user
	// what is being built rather than being a unique identifier.
	Name() string

	// Prepare configures the various components of this build and reports
	// any errors in doing so (such as syntax errors, validation errors, etc.)
	Prepare() error

	// Run runs the actual builder, returning an artifact implementation
	// of what is built. If anything goes wrong, an error is returned.
	Run(Ui, Cache) ([]Artifact, error)

	// Cancel will cancel a running build. This will block until the build
	// is actually completely cancelled.
	Cancel()

	// SetDebug will enable/disable debug mode. Debug mode is always
	// enabled by adding the additional key "packer_debug" to boolean
	// true in the configuration of the various components. This must
	// be called prior to Prepare.
	//
	// When SetDebug is set to true, parallelism between builds is
	// strictly prohibited.
	SetDebug(bool)
}

// A build struct represents a single build job, the result of which should
// be a single machine image artifact. This artifact may be comprised of
// multiple files, of course, but it should be for only a single provider
// (such as VirtualBox, EC2, etc.).
type coreBuild struct {
	name           string
	builder        Builder
	builderConfig  interface{}
	hooks          map[string][]Hook
	postProcessors [][]coreBuildPostProcessor
	provisioners   []coreBuildProvisioner

	debug         bool
	l             sync.Mutex
	prepareCalled bool
}

// Keeps track of the post-processor and the configuration of the
// post-processor used within a build.
type coreBuildPostProcessor struct {
	processor PostProcessor
	config    interface{}
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
	b.l.Lock()
	defer b.l.Unlock()

	if b.prepareCalled {
		panic("prepare already called")
	}

	b.prepareCalled = true

	debugConfig := map[string]interface{}{
		DebugConfigKey: b.debug,
	}

	// Prepare the builder
	err = b.builder.Prepare(b.builderConfig, debugConfig)
	if err != nil {
		log.Printf("Build '%s' prepare failure: %s\n", b.name, err)
		return
	}

	// Prepare the provisioners
	for _, coreProv := range b.provisioners {
		configs := make([]interface{}, len(coreProv.config), len(coreProv.config)+1)
		copy(configs, coreProv.config)
		configs = append(configs, debugConfig)

		if err = coreProv.provisioner.Prepare(configs...); err != nil {
			return
		}
	}

	// Prepare the post-processors
	for _, ppSeq := range b.postProcessors {
		for _, corePP := range ppSeq {
			if err = corePP.processor.Configure(corePP.config); err != nil {
				return
			}
		}
	}

	return
}

// Runs the actual build. Prepare must be called prior to running this.
func (b *coreBuild) Run(ui Ui, cache Cache) ([]Artifact, error) {
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
	artifacts := make([]Artifact, 0, 1)

	artifact, err := b.builder.Run(ui, hook, cache)
	if artifact != nil {
		artifacts = append(artifacts, artifact)
	}

	return artifacts, err
}

func (b *coreBuild) SetDebug(val bool) {
	if b.prepareCalled {
		panic("prepare has already been called")
	}

	b.debug = val
}

// Cancels the build if it is running.
func (b *coreBuild) Cancel() {
	b.builder.Cancel()
}
