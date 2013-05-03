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
	uiServer := &UiServer{ui}

	// Start the RPC server
	readyChan := make(chan int)
	stopChan := make(chan int)
	defer func() { stopChan <- 1 }()
	go testRPCServer(":1234", "Ui", uiServer, readyChan, stopChan)
	<-readyChan

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", ":1234")
	if err != nil {
		panic(err)
	}

	uiClient := &Ui{client}
	uiClient.Say("format", "arg0", 42)

	assert.Equal(ui.sayFormat, "format", "format should be correct")
}
