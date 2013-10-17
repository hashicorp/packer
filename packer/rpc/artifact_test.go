package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
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
	a := new(testArtifact)

	// Start the server
	server := rpc.NewServer()
	RegisterArtifact(server, a)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	aClient := Artifact(client)

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
