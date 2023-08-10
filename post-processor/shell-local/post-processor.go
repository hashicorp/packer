// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package shell_local

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	sl "github.com/hashicorp/packer-plugin-sdk/shell-local"
)

type PostProcessor struct {
	config sl.Config
}

type ExecuteCommandTemplate struct {
	Vars   string
	Script string
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := sl.Decode(&p.config, raws...)
	if err != nil {
		return err
	}
	if len(p.config.ExecuteCommand) == 1 {
		// Backwards compatibility -- before we merged the shell-local
		// post-processor and provisioners, the post-processor accepted
		// execute_command as a string rather than a slice of strings. It didn't
		// have a configurable call to shell program, automatically prepending
		// the user-supplied execute_command string with "sh -c". If users are
		// still using the old way of defining ExecuteCommand (by supplying a
		// single string rather than a slice of strings) then we need to
		// prepend this command with the call that the post-processor defaulted
		// to before.
		p.config.ExecuteCommand = append([]string{"sh", "-c"}, p.config.ExecuteCommand...)
	}

	return sl.Validate(&p.config)
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	generatedData := make(map[string]interface{})
	artifactStateData := artifact.State("generated_data")
	if artifactStateData != nil {
		for k, v := range artifactStateData.(map[interface{}]interface{}) {
			generatedData[k.(string)] = v
		}
	}

	success, retErr := sl.Run(ctx, ui, &p.config, generatedData)
	if !success {
		return nil, false, false, retErr
	}

	// Force shell-local pp to keep the input artifact, because otherwise we'll
	// lose it instead of being able to pass it through. If you want to delete
	// the input artifact for a shell local pp, use the artifice pp to create a
	// new artifact
	return artifact, true, true, retErr
}
