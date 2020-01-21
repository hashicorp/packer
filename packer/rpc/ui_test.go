package rpc

import (
	"bytes"
	"io"
	"io/ioutil"
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

	trackProgressCalled    bool
	progressBarAddCalled   bool
	progressBarCloseCalled bool
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

func (u *testUi) TrackProgress(_ string, _, _ int64, stream io.ReadCloser) (body io.ReadCloser) {
	u.trackProgressCalled = true

	return &readCloser{
		read: func(p []byte) (int, error) {
			u.progressBarAddCalled = true
			return stream.Read(p)
		},
		close: func() error {
			u.progressBarCloseCalled = true
			return stream.Close()
		},
	}
}

type readCloser struct {
	read  func([]byte) (int, error)
	close func() error
}

func (c *readCloser) Close() error               { return c.close() }
func (c *readCloser) Read(p []byte) (int, error) { return c.read(p) }

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

	ctt := []byte("foo bar baz !!!")
	rc := ioutil.NopCloser(bytes.NewReader(ctt))

	stream := uiClient.TrackProgress("stuff.txt", 0, int64(len(ctt)), rc)
	if ui.trackProgressCalled != true {
		t.Errorf("ProgressBastream not called.")
	}

	b := []byte{0}
	stream.Read(b) // output ignored
	if ui.progressBarAddCalled != true {
		t.Errorf("Add not called.")
	}

	stream.Close()
	if ui.progressBarCloseCalled != true {
		t.Errorf("close not called.")
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
