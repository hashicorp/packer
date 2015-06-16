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

const BuilderId = "packer.file"

type Builder struct {
	config *Config
	runner multistep.Runner
}

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

		// Create will truncate an existing file
		target, err := os.Create(b.config.Target)
		defer target.Close()
		if err != nil {
			return nil, err
		}

		ui.Say(fmt.Sprintf("Copying %s to %s", source.Name(), target.Name()))
		bytes, err := io.Copy(target, source)
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
