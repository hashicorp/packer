package terraformexport

import (
	"fmt"
	"os"
)

const BuilderId = "crunch.post-processor.terraform-export"

type Artifact struct {
	Variable	string
	Output		string
}

func NewArtifact(output, variable string) *Artifact {
	return &Artifact{
		Output:     output,
		Variable: 	variable,
	}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{a.Output}
}

func (a *Artifact) Id() string {
	return a.Variable
}

func (a *Artifact) String() string {
	return fmt.Sprintf("'%s' saved to: %s", a.Variable, a.Output)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return os.Remove(a.Output)
}
