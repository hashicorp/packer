package ncloud

import (
	"bytes"
	"fmt"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
)

const BuilderID = "ncloud.server.image"

type Artifact struct {
	MemberServerImage *server.MemberServerImage

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderID
}

func (a *Artifact) Files() []string {
	/* no file */
	return nil
}

func (a *Artifact) Id() string {
	return *a.MemberServerImage.MemberServerImageNo
}

func (a *Artifact) String() string {
	var buf bytes.Buffer

	// TODO : Logging artifact information
	buf.WriteString(fmt.Sprintf("%s:\n\n", a.BuilderId()))
	buf.WriteString(fmt.Sprintf("Member Server Image Name: %s\n", *a.MemberServerImage.MemberServerImageName))
	buf.WriteString(fmt.Sprintf("Member Server Image No: %s\n", *a.MemberServerImage.MemberServerImageNo))

	return buf.String()
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}
	return a.MemberServerImage.MemberServerImageStatus
}

func (a *Artifact) Destroy() error {
	return nil
}
