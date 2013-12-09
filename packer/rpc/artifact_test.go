package rpc

import (
	"github.com/mitchellh/packer/packer"
	"reflect"
	"testing"
)

type testArtifact struct{}

func (testArtifact) BuilderId() string {
	return "bid"
}

func (testArtifact) Files() []string {
	return []string{"a", "b"}
}

func (testArtifact) Id() string {
	return "id"
}

func (testArtifact) String() string {
	return "string"
}

func (testArtifact) Destroy() error {
	return nil
}

func TestArtifactRPC(t *testing.T) {
	// Create the interface to test
	a := new(packer.MockArtifact)

	// Start the server
	server := NewServer()
	server.RegisterArtifact(a)
	client := testClient(t, server)
	aClient := client.Artifact()

	// Test
	if aClient.BuilderId() != "bid" {
		t.Fatalf("bad: %s", aClient.BuilderId())
	}

	if !reflect.DeepEqual(aClient.Files(), []string{"a", "b"}) {
		t.Fatalf("bad: %#v", aClient.Files())
	}

	if aClient.Id() != "id" {
		t.Fatalf("bad: %s", aClient.Id())
	}

	if aClient.String() != "string" {
		t.Fatalf("bad: %s", aClient.String())
	}
}

func TestArtifact_Implements(t *testing.T) {
	var _ packer.Artifact = Artifact(nil)
}
