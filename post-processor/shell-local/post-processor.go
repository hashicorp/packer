package shell_local

import (
	"runtime"

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

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := sl.Decode(&p.config, raws...)
	if err != nil {
		return err
	}
	if len(p.config.ExecuteCommand) == 0 && runtime.GOOS != "windows" {
		// Backwards compatibility from before post-processor merge with
		// provisioner. Don't need to default separately for windows becuase the
		// post-processor never worked for windows before the merge with the
		// provisioner code, so the provisioner defaults are fine.
		p.config.ExecuteCommand = []string{"sh", "-c", `chmod +x "{{.Script}}"; {{.Vars}} "{{.Script}}"`}
	} else if len(p.config.ExecuteCommand) == 1 {
		// Backwards compatibility -- before merge, post-processor didn't have
		// configurable call to shell program, meaning users may not have
		// defined this in their call. If users are still using the old way of
		// defining ExecuteCommand (e.g. just supplying a single string that is
		// now being interpolated as a slice with one item), then assume we need
		// to prepend this call still, and use the one that the post-processor
		// defaulted to before.
		p.config.ExecuteCommand = append([]string{"sh", "-c"}, p.config.ExecuteCommand...)
	}

	return sl.Validate(&p.config)
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	// this particular post-processor doesn't do anything with the artifact
	// except to return it.

	retBool, retErr := sl.Run(ui, &p.config)
	if !retBool {
		return nil, retBool, retErr
	}

	return artifact, retBool, retErr
}
