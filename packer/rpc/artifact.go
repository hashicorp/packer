package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.Artifact where the artifact is actually
// available over an RPC connection.
type artifact struct {
	client *rpc.Client
}

// ArtifactServer wraps a packer.Artifact implementation and makes it
// exportable as part of a Golang RPC server.
type ArtifactServer struct {
	artifact packer.Artifact
}

func Artifact(client *rpc.Client) *artifact {
	return &artifact{client}
}

func (a *artifact) BuilderId() (result string) {
	a.client.Call("Artifact.BuilderId", new(interface{}), &result)
	return
}

func (a *artifact) Files() (result []string) {
	a.client.Call("Artifact.Files", new(interface{}), &result)
	return
}

func (a *artifact) Id() (result string) {
	a.client.Call("Artifact.Id", new(interface{}), &result)
	return
}

func (a *artifact) String() (result string) {
	a.client.Call("Artifact.String", new(interface{}), &result)
	return
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
