package manifest

import "fmt"

const BuilderId = "packer.post-processor.manifest"

type Artifact struct {
	BuildName   string   `json:"build_name"`
	BuildTime   int64    `json:"build_time"`
	Description string   `json:"description"`
	BuildFiles  []string `json:"files"`
	BuildId     string   `json:"artifact_id"`
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.BuildFiles
}

func (a *Artifact) Id() string {
	return a.BuildId
}

func (a *Artifact) String() string {
	return fmt.Sprintf("%s-%s", a.BuildName, a.BuildId)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
