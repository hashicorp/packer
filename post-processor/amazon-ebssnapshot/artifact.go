package ebssnapshot

import (
	"fmt"
	"sort"
	"strings"
)

const BuilderId = "packer.post-processor.amazon-ebssnapshot"

type Artifact struct {
	Snapshots []string
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	// We have no files
	return nil
}

// returns a sorted list of snapshot IDs
func (a *Artifact) idList() []string {
	sort.Strings(a.Snapshots)
	return a.Snapshots
}

func (a *Artifact) Id() string {
	return strings.Join(a.idList(), ",")
}

func (a *Artifact) String() string {
	return fmt.Sprintf("EBS Volume Snapshots were created:\n\n%s", strings.Join(a.idList(), "\n"))
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
