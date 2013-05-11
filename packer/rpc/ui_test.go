package rpc

import (
	"cgl.tideland.biz/asserts"
	"net/rpc"
	"testing"
)

type testUi struct {
	errorCalled bool
	errorFormat string
	errorVars   []interface{}
	sayCalled   bool
	sayFormat   string
	sayVars     []interface{}
}

func (u *testUi) Error(format string, a ...interface{}) {
	u.errorCalled = true
	u.errorFormat = format
	u.errorVars = a
}

func (u *testUi) Say(format string, a ...interface{}) {
	u.sayCalled = true
	u.sayFormat = format
	u.sayVars = a
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

	uiClient.Error("format", "arg0", 42)
	assert.Equal(ui.errorFormat, "format", "format should be correct")

	uiClient.Say("format", "arg0", 42)
	assert.Equal(ui.sayFormat, "format", "format should be correct")
}
