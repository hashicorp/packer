package rpc

import (
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// An implementation of packersdk.Artifact where the artifact is actually
// available over an RPC connection.
type artifact struct {
	commonClient
}

// ArtifactServer wraps a packersdk.Artifact implementation and makes it
// exportable as part of a Golang RPC server.
type ArtifactServer struct {
	artifact packersdk.Artifact
}

func (a *artifact) BuilderId() (result string) {
	a.client.Call(a.endpoint+".BuilderId", new(interface{}), &result)
	return
}

func (a *artifact) Files() (result []string) {
	a.client.Call(a.endpoint+".Files", new(interface{}), &result)
	return
}

func (a *artifact) Id() (result string) {
	a.client.Call(a.endpoint+".Id", new(interface{}), &result)
	return
}

func (a *artifact) String() (result string) {
	a.client.Call(a.endpoint+".String", new(interface{}), &result)
	return
}

func (a *artifact) State(name string) (result interface{}) {
	a.client.Call(a.endpoint+".State", name, &result)
	return
}

func (a *artifact) Destroy() error {
	var result error
	if err := a.client.Call(a.endpoint+".Destroy", new(interface{}), &result); err != nil {
		return err
	}

	return result
}

func (s *ArtifactServer) BuilderId(args *interface{}, reply *string) error {
	*reply = s.artifact.BuilderId()
	return nil
}

func (s *ArtifactServer) Files(args *interface{}, reply *[]string) error {
	*reply = s.artifact.Files()
	return nil
}

func (s *ArtifactServer) Id(args *interface{}, reply *string) error {
	*reply = s.artifact.Id()
	return nil
}

func (s *ArtifactServer) String(args *interface{}, reply *string) error {
	*reply = s.artifact.String()
	return nil
}

func (s *ArtifactServer) State(name string, reply *interface{}) error {
	*reply = s.artifact.State(name)
	return nil
}

func (s *ArtifactServer) Destroy(args *interface{}, reply *error) error {
	err := s.artifact.Destroy()
	if err != nil {
		err = NewBasicError(err)
	}

	*reply = err
	return nil
}
