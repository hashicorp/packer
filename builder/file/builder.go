package file

/*
The File builder creates an artifact from a file. Because it does not require
any virutalization or network resources, it's very fast and useful for testing.
*/

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

const BuilderId = "cbednarski.file"

type Builder struct {
	config *Config
	runner multistep.Runner
}

// Prepare is responsible for configuring the builder and validating
// that configuration. Any setup should be done in this method. Note that
// NO side effects should take place in prepare, it is meant as a state
// setup only. Calling Prepare is not necessarilly followed by a Run.
//
// The parameters to Prepare are a set of interface{} values of the
// configuration. These are almost always `map[string]interface{}`
// parsed from a template, but no guarantee is made.
//
// Each of the configuration values should merge into the final
// configuration.
//
// Prepare should return a list of warnings along with any errors
// that occured while preparing.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	return warnings, nil
}

// Run is where the actual build should take place. It takes a Build and a Ui.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	artifact := new(FileArtifact)

	if b.config.Source != "" {
		source, err := os.Open(b.config.Source)
		defer source.Close()
		if err != nil {
			return nil, err
		}

		target, err := os.OpenFile(b.config.Target, os.O_WRONLY, 0600)
		defer target.Close()
		if err != nil {
			return nil, err
		}

		ui.Say(fmt.Sprintf("Copying %s to %s", source.Name(), target.Name()))
		bytes, err := io.Copy(source, target)
		if err != nil {
			return nil, err
		}
		ui.Say(fmt.Sprintf("Copied %d bytes", bytes))
		artifact.filename = target.Name()
	} else {
		// We're going to write Contents; if it's empty we'll just create an
		// empty file.
		err := ioutil.WriteFile(b.config.Target, []byte(b.config.Content), 0600)
		if err != nil {
			return nil, err
		}
		artifact.filename = b.config.Target
	}

	return artifact, nil
}

// Cancel cancels a possibly running Builder. This should block until
// the builder actually cancels and cleans up after itself.
func (b *Builder) Cancel() {
	b.runner.Cancel()
}
