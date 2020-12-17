package packer

import (
	"context"
)

// Implementers of Builder are responsible for actually building images
// on some platform given some configuration.
//
// In addition to the documentation on Prepare above: Prepare is sometimes
// configured with a `map[string]interface{}` that has a key "packer_debug".
// This is a boolean value. If it is set to true, then the builder should
// enable a debug mode which allows builder developers and advanced users
// to introspect what is going on during a build. During debug builds,
// parallelism is strictly disabled, so it is safe to request input from
// stdin and so on.
type Builder interface {
	HCL2Speccer

	// Prepare is responsible for configuring the builder and validating
	// that configuration. Any setup should be done in this method. Note that
	// NO side effects should take place in prepare, it is meant as a state
	// setup only. Calling Prepare is not necessarily followed by a Run.
	//
	// The parameters to Prepare are a set of interface{} values of the
	// configuration. These are almost always `map[string]interface{}`
	// parsed from a template, but no guarantee is made.
	//
	// Each of the configuration values should merge into the final
	// configuration.
	//
	// Prepare should return a list of variables that will be made accessible to
	// users during the provision methods, a list of warnings along with any
	// errors that occurred while preparing.
	Prepare(...interface{}) ([]string, []string, error)

	// Run is where the actual build should take place. It takes a Build and a Ui.
	Run(context.Context, Ui, Hook) (Artifact, error)
}
