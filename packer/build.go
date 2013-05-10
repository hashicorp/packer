package packer

import "log"

// A Build represents a single job within Packer that is responsible for
// building some machine image artifact. Builds are meant to be parallelized.
type Build interface {
	Name() string
	Prepare() error
	Run(ui Ui)
}

// A build struct represents a single build job, the result of which should
// be a single machine image artifact. This artifact may be comprised of
// multiple files, of course, but it should be for only a single provider
// (such as VirtualBox, EC2, etc.).
type coreBuild struct {
	name    string
	builder Builder
	rawConfig interface{}

	prepareCalled bool
}

// Returns the name of the build.
func (b *coreBuild) Name() string {
	return b.name
}

// Prepare prepares the build by doing some initialization for the builder
// and any hooks. This _must_ be called prior to Run.
func (b *coreBuild) Prepare() (err error) {
	b.prepareCalled = true
	err = b.builder.Prepare(b.rawConfig)
	if err != nil {
		log.Printf("Build '%s' prepare failure: %s\n", b.name, err)
	}

	return
}

// Runs the actual build. Prepare must be called prior to running this.
func (b *coreBuild) Run(ui Ui) {
	if !b.prepareCalled {
		panic("Prepare must be called first")
	}

	b.builder.Run(b, ui)
}
