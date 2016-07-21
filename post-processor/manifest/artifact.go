package manifest

import "fmt"

const BuilderId = "packer.post-processor.manifest"

type Artifact struct {
	BuildName     string   `json:"name"`
	BuilderType   string   `json:"builder_type"`
	BuildTime     int64    `json:"build_time"`
	ArtifactFiles []string `json:"files"`
	ArtifactId    string   `json:"artifact_id"`
	PackerRunUUID string   `json:"packer_run_uuid"`
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.ArtifactFiles
}

func (a *Artifact) Id() string {
	return a.ArtifactId
}

func (a *Artifact) String() string {
	return fmt.Sprintf("%s-%s", a.BuildName, a.ArtifactId)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
