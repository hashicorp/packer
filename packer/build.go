package packer

// A build struct represents a single build job, the result of which should
// be a single machine image artifact. This artifact may be comprised of
// multiple files, of course, but it should be for only a single provider
// (such as VirtualBox, EC2, etc.).
type Build struct {
	name    string
	builder Builder
	rawConfig interface{}

	prepareCalled bool
}

// Implementers of Builder are responsible for actually building images
// on some platform given some configuration.
//
// Prepare is responsible for reading in some configuration, in the raw form
// of map[string]interface{}, and storing that state for use later. Any setup
// should be done in this method. Note that NO side effects should really take
// place in prepare. It is meant as a state setup step only.
//
// Run is where the actual build should take place. It takes a Build and a Ui.
type Builder interface {
	Prepare(config interface{})
	Run(build *Build, ui Ui)
}

// This factory is responsible for returning Builders for the given name.
//
// CreateBuilder is called with the string name of a builder and should
// return a Builder implementation by that name or nil if no builder can be
// found.
type BuilderFactory interface {
	CreateBuilder(name string) Builder
}

// This implements BuilderFactory to return nil for every builder.
type NilBuilderFactory byte

func (NilBuilderFactory) CreateBuilder(name string) Builder {
	return nil
}

// Prepare prepares the build by doing some initialization for the builder
// and any hooks. This _must_ be called prior to Run.
func (b *Build) Prepare() {
	b.prepareCalled = true
	b.builder.Prepare(b.rawConfig)
}

// Runs the actual build. Prepare must be called prior to running this.
func (b *Build) Run(ui Ui) {
	if !b.prepareCalled {
		panic("Prepare must be called first")
	}

	b.builder.Run(b, ui)
}
