package packer

// A PreProcessor is responsible for taking an artifact of a build
// and doing some sort of pre-processing to turn this into another
// artifact. An example of a pre-processor would be something that takes
// the result of a build, compresses it, and returns a new artifact containing
// a single file of the prior artifact compressed.
type PreProcessor interface {
	// Configure is responsible for setting up configuration, storing
	// the state for later, and returning and errors, such as validation
	// errors.
	Configure(...interface{}) error

	// PreProcess is called to run tasks before builders.
	// If an error occurs, it should return that error.
	PreProcess(Ui) error
}
