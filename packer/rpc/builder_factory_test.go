package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

var createResult = &testBuilder{}

type testBuilderFactory struct {
	createCalled bool
	createName string
}

func (b *testBuilderFactory) CreateBuilder(name string) packer.Builder {
	b.createCalled = true
	b.createName = name
	return createResult
}

func TestBuilderFactoryRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the interface to test
	b := new(testBuilderFactory)

	// Start the server
	server := NewServer()
	server.RegisterBuilderFactory(b)
	server.Start()
	defer server.Stop()

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", server.Address())
	assert.Nil(err, "should be able to connect")

	// Test Create
	name := "foo"
	bClient := &BuilderFactory{client}
	builder := bClient.CreateBuilder(name)
	assert.True(b.createCalled, "create should be called")
	assert.Equal(b.createName, "foo", "name should be foo")

	builder.Prepare(42)
	assert.True(createResult.prepareCalled, "prepare should be called")
	assert.Equal(createResult.prepareConfig, 42, "42 should be config")
}

func TestBuilderFactory_ImplementsBuilderFactory(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var realVar packer.BuilderFactory
	b := &BuilderFactory{nil}

	assert.Implementor(b, &realVar, "should be a BuilderFactory")
}
