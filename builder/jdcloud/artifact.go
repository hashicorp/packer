package jdcloud

import (
	"fmt"
	"sort"
	"strings"
)

type Artifact struct {
	ImageId  string
	RegionID string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BUILDER_ID
}

func (*Artifact) Files() []string {
	return nil
}

// Plan
// Though this part is supposed to be an array of Image Ids associated
// with its region, but currently only a single image is supported
func (a *Artifact) Id() string {
	parts := []string{fmt.Sprintf("%s:%s", a.RegionID, a.ImageId)}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A VMImage was created: %s", a.ImageId)
}

// Plan
// State and destroy function is abandoned
func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	return nil
}
