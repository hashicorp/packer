package common

import (
	"fmt"
	"log"
	"sort"
	"strings"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

type Artifact struct {
	UCloudImages *ImageInfoSet

	BuilderIdValue string

	Client *UCloudClient

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	m := make([]string, 0, len(a.UCloudImages.GetAll()))

	for _, v := range a.UCloudImages.GetAll() {
		m = append(m, fmt.Sprintf("%s:%s:%s", v.ProjectId, v.Region, v.ImageId))
	}

	sort.Strings(m)
	return strings.Join(m, ",")
}

func (a *Artifact) String() string {
	m := make([]string, 0, len(a.UCloudImages.GetAll()))
	for _, v := range a.UCloudImages.GetAll() {
		m = append(m, fmt.Sprintf("%s: %s: %s", v.ProjectId, v.Region, v.ImageId))
	}

	sort.Strings(m)
	return fmt.Sprintf("UCloud images were created:\n\n%s", strings.Join(m, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}

	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {
	conn := a.Client.UHostConn
	errors := make([]error, 0)

	for _, v := range a.UCloudImages.GetAll() {
		log.Printf("Delete ucloud image %s from %s:%s", v.ImageId, v.ProjectId, v.Region)
		req := conn.NewTerminateCustomImageRequest()
		req.ProjectId = ucloud.String(v.ProjectId)
		req.Region = ucloud.String(v.Region)
		req.ImageId = ucloud.String(v.ImageId)

		if _, err := conn.TerminateCustomImage(req); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		} else {
			return &packersdk.MultiError{Errors: errors}
		}
	}

	return nil
}

func (a *Artifact) stateAtlasMetadata() interface{} {
	metadata := make(map[string]string)
	for _, v := range a.UCloudImages.GetAll() {
		k := fmt.Sprintf("%s:%s", v.ProjectId, v.Region)
		metadata[k] = v.ImageId
	}

	return metadata
}
