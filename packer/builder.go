package packer

// Implementers of Builder are responsible for actually building images
// on some platform given some configuration.
type Builder interface {
	// Prepare is responsible for reading in some configuration, in the raw form
	// of map[string]interface{}, and storing that state for use later. Any setup
	// should be done in this method. Note that NO side effects should really take
	// place in prepare. It is meant as a state setup step only.
	Prepare(config interface{}) error

	// Run is where the actual build should take place. It takes a Build and a Ui.
	Run(ui Ui, hook Hook) Artifact

	// Cancel cancels a possibly running Builder. This should block until
	// the builder actually cancels and cleans up after itself.
	Cancel()
}
