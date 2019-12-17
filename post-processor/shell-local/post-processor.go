package shell_local

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/packer"
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

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	// this particular post-processor doesn't do anything with the artifact
	// except to return it.

	success, retErr := sl.Run(ctx, ui, &p.config)
	if !success {
		return nil, false, false, retErr
	}

	// Force shell-local pp to keep the input artifact, because otherwise we'll
	// lose it instead of being able to pass it through. If you want to delete
	// the input artifact for a shell local pp, use the artifice pp to create a
	// new artifact
	return artifact, true, true, retErr
}
