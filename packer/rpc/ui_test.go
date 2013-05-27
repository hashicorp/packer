package rpc

import (
	"cgl.tideland.biz/asserts"
	"net/rpc"
	"testing"
)

type testUi struct {
	errorCalled bool
	errorMessage string
	sayCalled   bool
	sayMessage   string
}

func (u *testUi) Error(message string) {
	u.errorCalled = true
	u.errorMessage = message
}

func (u *testUi) Say(message string) {
	u.sayCalled = true
	u.sayMessage = message
}

func TestUiRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the UI to test
	ui := new(testUi)

	// Start the RPC server
	server := rpc.NewServer()
	RegisterUi(server, ui)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	uiClient := &Ui{client}

	// Basic error and say tests
	uiClient.Error("message")
	assert.Equal(ui.errorMessage, "message", "message should be correct")

	uiClient.Say("message")
	assert.Equal(ui.sayMessage, "message", "message should be correct")
}
