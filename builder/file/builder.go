package file

/*
The File builder creates an artifact from a file. Because it does not require
any virtualization or network resources, it's very fast and useful for testing.
*/

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
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
