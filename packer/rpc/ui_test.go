package rpc

import (
	"cgl.tideland.biz/asserts"
	"net/rpc"
	"testing"
)

type testUi struct {
	askCalled bool
	askQuery string
	errorCalled    bool
	errorMessage   string
	messageCalled  bool
	messageMessage string
	sayCalled      bool
	sayMessage     string
}

func (u *testUi) Ask(query string) string {
	u.askCalled = true
	u.askQuery = query
	return "foo"
}

func (u *testUi) Error(message string) {
	u.errorCalled = true
	u.errorMessage = message
}

func (u *testUi) Message(message string) {
	u.messageCalled = true
	u.messageMessage = message
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
	result := uiClient.Ask("query")
	assert.True(ui.askCalled, "ask should be called")
	assert.Equal(ui.askQuery, "query", "should be correct")
	assert.Equal(result, "foo", "should have correct result")

	uiClient.Error("message")
	assert.Equal(ui.errorMessage, "message", "message should be correct")

	uiClient.Message("message")
	assert.Equal(ui.messageMessage, "message", "message should be correct")

	uiClient.Say("message")
	assert.Equal(ui.sayMessage, "message", "message should be correct")
}
