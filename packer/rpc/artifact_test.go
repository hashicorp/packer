package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
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

func TestArtifactRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the interface to test
	a := new(testArtifact)

	// Start the server
	server := rpc.NewServer()
	RegisterArtifact(server, a)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	assert.Nil(err, "should be able to connect")
	aClient := Artifact(client)

	// Test
	assert.Equal(aClient.BuilderId(), "bid", "should have correct builder ID")
	assert.Equal(aClient.Files(), []string{"a", "b"}, "should have correct builder ID")
	assert.Equal(aClient.Id(), "id", "should have correct builder ID")
	assert.Equal(aClient.String(), "string", "should have correct builder ID")
}

func TestArtifact_Implements(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r packer.Artifact
	a := Artifact(nil)

	assert.Implementor(a, &r, "should be an Artifact")
}
