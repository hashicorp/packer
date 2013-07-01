package packer

// A PostProcessor is responsible for taking an artifact of a build
// and doing some sort of post-processing to turn this into another
// artifact. An example of a post-processor would be something that takes
// the result of a build, compresses it, and returns a new artifact containing
// a single file of the prior artifact compressed.
type PostProcessor interface {
	// Configure is responsible for setting up configuration, storing
	// the state for later, and returning and errors, such as validation
	// errors.
	Configure(...interface{}) error

	// PostProcess takes a previously created Artifact and produces another
	// Artifact. If an error occurs, it should return that error. If `keep`
	// is to true, then the previous artifact is forcibly kept.
	PostProcess(Ui, Artifact) (a Artifact, keep bool, err error)
}
