package rpc

import (
	"net/rpc"
	"reflect"
	"testing"
)

type testUi struct {
	askCalled      bool
	askQuery       string
	errorCalled    bool
	errorMessage   string
	machineCalled  bool
	machineType    string
	machineArgs    []string
	messageCalled  bool
	messageMessage string
	sayCalled      bool
	sayMessage     string
}

func (u *testUi) Ask(query string) (string, error) {
	u.askCalled = true
	u.askQuery = query
	return "foo", nil
}

func (u *testUi) Error(message string) {
	u.errorCalled = true
	u.errorMessage = message
}

func (u *testUi) Machine(t string, args ...string) {
	u.machineCalled = true
	u.machineType = t
	u.machineArgs = args
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

	uiClient := &Ui{client: client, endpoint: "Ui"}

	// Basic error and say tests
	result, err := uiClient.Ask("query")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if !ui.askCalled {
		t.Fatal("should be called")
	}
	if ui.askQuery != "query" {
		t.Fatalf("bad: %s", ui.askQuery)
	}
	if result != "foo" {
		t.Fatalf("bad: %#v", result)
	}

	uiClient.Error("message")
	if ui.errorMessage != "message" {
		t.Fatalf("bad: %#v", ui.errorMessage)
	}

	uiClient.Message("message")
	if ui.messageMessage != "message" {
		t.Fatalf("bad: %#v", ui.errorMessage)
	}

	uiClient.Say("message")
	if ui.sayMessage != "message" {
		t.Fatalf("bad: %#v", ui.errorMessage)
	}

	uiClient.Machine("foo", "bar", "baz")
	if !ui.machineCalled {
		t.Fatal("machine should be called")
	}

	if ui.machineType != "foo" {
		t.Fatalf("bad type: %#v", ui.machineType)
	}

	expected := []string{"bar", "baz"}
	if !reflect.DeepEqual(ui.machineArgs, expected) {
		t.Fatalf("bad: %#v", ui.machineArgs)
	}
}
