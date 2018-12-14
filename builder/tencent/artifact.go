package tencent

import (
	"fmt"
	"log"
)

// Artifact is an artifact implementation that contains built Tencent images.
type Artifact struct {
	// BuilderIDValue is the unique ID for the builder that created this Image
	BuilderIDValue string

	Config Config
	Driver Driver

	// InstanceId of the artifact
	InstanceId     string
	SSHKeyLocation string
	IPAddress      string
}

// BuilderId must be returned
func (a *Artifact) BuilderId() string {
	return a.BuilderIDValue
}

// Files returns the files used/required by the artifact
func (a *Artifact) Files() []string {
	return []string{a.SSHKeyLocation}
}

// Id returns the ID of the artifact
func (a *Artifact) Id() string {
	return a.InstanceId
}

// String returns the ID of the artifact
func (a *Artifact) String() string {
	var msg string
	if a.InstanceId != "" {
		msg = fmt.Sprintf("Instance was created: %s", a.InstanceId)
		if a.IPAddress != "" {
			msg = fmt.Sprintf("%s and IP address is: %s", msg, a.IPAddress)
		}
	} else {
		msg = "No instance was created"
	}
	return msg
}

// State returns the state specified in the artifact
func (a *Artifact) State(name string) interface{} {
	switch name {
	case CArtifactIPAddress:
		return a.IPAddress
	case CArtifactBuilderID:
		return a.BuilderIDValue
	case CInstanceId:
		return a.InstanceId
	default:
		return nil
	}
}

// Destroy destroys the image specified by the artifact
func (a *Artifact) Destroy() error {
	log.Printf("Deleting instance id: %s", a.InstanceId)

	return nil
}
