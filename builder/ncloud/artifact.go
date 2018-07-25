package ncloud

import (
	"bytes"
	"fmt"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
)

const BuilderID = "ncloud.server.image"

type Artifact struct {
	ServerImage *ncloud.ServerImage
}

func (*Artifact) BuilderId() string {
	return BuilderID
}

func (a *Artifact) Files() []string {
	/* no file */
	return nil
}

func (a *Artifact) Id() string {
	return a.ServerImage.MemberServerImageNo
}

func (a *Artifact) String() string {
	var buf bytes.Buffer

	// TODO : Logging artifact information
	buf.WriteString(fmt.Sprintf("%s:\n\n", a.BuilderId()))
	buf.WriteString(fmt.Sprintf("Member Server Image Name: %s\n", a.ServerImage.MemberServerImageName))
	buf.WriteString(fmt.Sprintf("Member Server Image No: %s\n", a.ServerImage.MemberServerImageNo))

	return buf.String()
}

func (a *Artifact) State(name string) interface{} {
	return a.ServerImage.MemberServerImageStatus
}

func (a *Artifact) Destroy() error {
	return nil
}
