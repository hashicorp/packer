package rpc

import (
	"io"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer"
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

	progressBarCalled               bool
	progressBarStartCalled          bool
	progressBarAddCalled            bool
	progressBarFinishCalled         bool
	progressBarNewProxyReaderCalled bool
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

func (u *testUi) ProgressBar(string) packer.ProgressBar {
	u.progressBarCalled = true
	return u
}

func (u *testUi) Start(int64) {
	u.progressBarStartCalled = true
}

func (u *testUi) Add(int64) {
	u.progressBarAddCalled = true
}

func (u *testUi) Finish() {
	u.progressBarFinishCalled = true
}

func (u *testUi) NewProxyReader(r io.Reader) io.Reader {
	u.progressBarNewProxyReaderCalled = true
	return r
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

	bar := uiClient.ProgressBar("test")
	if ui.progressBarCalled != true {
		t.Errorf("ProgressBar not called.")
	}

	bar.Start(100)
	if ui.progressBarStartCalled != true {
		t.Errorf("progressBar.Start not called.")
	}

	bar.Add(1)
	if ui.progressBarAddCalled != true {
		t.Errorf("progressBar.Add not called.")
	}

	bar.Finish()
	if ui.progressBarFinishCalled != true {
		t.Errorf("progressBar.Finish not called.")
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
