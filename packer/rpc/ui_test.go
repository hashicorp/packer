package rpc

import (
	"github.com/hashicorp/packer/packer"
	"reflect"
	"testing"
)

type testUi struct {
	askCalled                   bool
	askQuery                    string
	errorCalled                 bool
	errorMessage                string
	machineCalled               bool
	machineType                 string
	machineArgs                 []string
	messageCalled               bool
	messageMessage              string
	sayCalled                   bool
	sayMessage                  string
	getProgressBarCalled        bool
	getProgressBarValue         packer.ProgressBar
	progessBarCallbackWasCalled bool
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

func (u *testUi) GetProgressBar() packer.ProgressBar {
	u.getProgressBarCalled = true
	u.getProgressBarValue = packer.GetDummyProgressBar()
	u.getProgressBarValue.Callback = func(string) {
		u.progessBarCallbackWasCalled = true
	}
	return u.getProgressBarValue
}

func TestUiRPC(t *testing.T) {
	// Create the UI to test
	ui := new(testUi)

	// Start the RPC server
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterUi(ui)

	uiClient := client.Ui()

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

	bar := uiClient.GetProgressBar()
	if !ui.getProgressBarCalled {
		t.Fatal("getprogressbar should be called")
	}

	if bar.Callback == nil {
		t.Fatal("getprogressbar returned a bar with an empty callback")
	}
	bar.Callback("test")
	if !ui.progessBarCallbackWasCalled {
		t.Fatal("progressbarcallback should be called")
	}

	if ui.machineType != "foo" {
		t.Fatalf("bad type: %#v", ui.machineType)
	}

	expected := []string{"bar", "baz"}
	if !reflect.DeepEqual(ui.machineArgs, expected) {
		t.Fatalf("bad: %#v", ui.machineArgs)
	}
}
