package packer

// A build struct represents a single build job, the result of which should
// be a single machine image artifact. This artifact may be comprised of
// multiple files, of course, but it should be for only a single provider
// (such as VirtualBox, EC2, etc.).
type Build struct {
	name    string
	builder Builder
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
	Run(build Build, ui Ui)
}

// This factory is responsible for returning Builders for the given name.
//
// CreateBuilder is called with the string name of a builder and should
// return a Builder implementation by that name or nil if no builder can be
// found.
type BuilderFactory interface {
	CreateBuilder(name string) Builder
}
