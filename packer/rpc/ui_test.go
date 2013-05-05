package rpc

import (
	"cgl.tideland.biz/asserts"
	"net/rpc"
	"testing"
)

type testUi struct {
	sayCalled bool
	sayFormat string
	sayVars []interface{}
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
	server := NewServer()
	server.RegisterUi(ui)
	server.Start()
	defer server.Stop()

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", server.Address())
	if err != nil {
		panic(err)
	}

	uiClient := &Ui{client}
	uiClient.Say("format", "arg0", 42)

	assert.Equal(ui.sayFormat, "format", "format should be correct")
}
