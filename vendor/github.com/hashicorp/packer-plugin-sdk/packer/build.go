package packer

import "context"

// A Build represents a single job within Packer that is responsible for
// building some machine image artifact. Builds are meant to be parallelized.
type Build interface {
	// Name is the name of the build. This is unique across a single template,
	// but not absolutely unique. This is meant more to describe to the user
	// what is being built rather than being a unique identifier.
	Name() string

	// Prepare configures the various components of this build and reports
	// any errors in doing so (such as syntax errors, validation errors, etc.).
	// It also reports any warnings.
	Prepare() ([]string, error)

	// Run runs the actual builder, returning an artifact implementation
	// of what is built. If anything goes wrong, an error is returned.
	// Run can be context cancelled.
	Run(context.Context, Ui) ([]Artifact, error)

	// SetDebug will enable/disable debug mode. Debug mode is always
	// enabled by adding the additional key "packer_debug" to boolean
	// true in the configuration of the various components. This must
	// be called prior to Prepare.
	//
	// When SetDebug is set to true, parallelism between builds is
	// strictly prohibited.
	SetDebug(bool)

	// SetForce will enable/disable forcing a build when artifacts exist.
	//
	// When SetForce is set to true, existing artifacts from the build are
	// deleted prior to the build.
	SetForce(bool)

	// SetOnError will determine what to do when a normal multistep step fails
	// - "cleanup" - run cleanup steps
	// - "abort" - exit without cleanup
	// - "ask" - ask the user
	SetOnError(string)
}
