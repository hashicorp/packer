package packer

import (
	"context"
)

// A PostProcessor is responsible for taking an artifact of a build
// and doing some sort of post-processing to turn this into another
// artifact. An example of a post-processor would be something that takes
// the result of a build, compresses it, and returns a new artifact containing
// a single file of the prior artifact compressed.
type PostProcessor interface {
	HCL2Speccer

	// Configure is responsible for setting up configuration, storing
	// the state for later, and returning and errors, such as validation
	// errors.
	Configure(...interface{}) error

	// PostProcess takes a previously created Artifact and produces another
	// Artifact. If an error occurs, it should return that error. If `keep` is
	// true, then the previous artifact defaults to being kept if user has not
	// given a value to keep_input_artifact. If forceOverride is true, then any
	// user input for keep_input_artifact is ignored and the artifact is either
	// kept or discarded according to the value set in `keep`.
	// PostProcess is cancellable using context
	PostProcess(context.Context, Ui, Artifact) (a Artifact, keep bool, forceOverride bool, err error)
}
