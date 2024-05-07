// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package file

/*
The File builder creates an artifact from a file. Because it does not require
any virtualization or network resources, it's very fast and useful for testing.
*/

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const BuilderId = "packer.file"

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

// Run is where the actual build should take place. It takes a Build and a Ui.
func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	artifact := new(FileArtifact)

	// Create all directories leading to target
	dir := filepath.Dir(b.config.Target)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	if b.config.Source != "" {
		source, err := os.Open(b.config.Source)
		if err != nil {
			return nil, err
		}
		defer source.Close()

		// Create will truncate an existing file
		target, err := os.Create(b.config.Target)
		if err != nil {
			return nil, err
		}
		defer target.Close()

		ui.Say(fmt.Sprintf("Copying %s to %s", source.Name(), target.Name()))
		bytes, err := io.Copy(target, source)
		if err != nil {
			return nil, err
		}
		ui.Say(fmt.Sprintf("Copied %d bytes", bytes))

		artifact.source = b.config.Source
		artifact.filename = target.Name()
	} else {
		// We're going to write Contents; if it's empty we'll just create an
		// empty file.
		err := os.WriteFile(b.config.Target, []byte(b.config.Content), 0600)
		if err != nil {
			return nil, err
		}
		artifact.source = "<no-defined-source-file>"
		artifact.filename = b.config.Target
	}

	if hook != nil {
		if err := hook.Run(ctx, packersdk.HookProvision, ui, new(packersdk.MockCommunicator), nil); err != nil {
			return nil, err
		}
	}

	return artifact, nil
}
