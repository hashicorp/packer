package common

import (
	"github.com/hashicorp/packer/packer"
	"testing"
)

// test packer.Ui implementation to verify that progress bar is being written
type testProgressBarUi struct {
	messageCalled  bool
	messageMessage string
}

func (u *testProgressBarUi) Say(string)                {}
func (u *testProgressBarUi) Error(string)              {}
func (u *testProgressBarUi) Machine(string, ...string) {}

func (u *testProgressBarUi) Ask(string) (string, error) {
	return "", nil
}
func (u *testProgressBarUi) Message(message string) {
	u.messageCalled = true
	u.messageMessage = message
}

// ..and now let's begin our actual tests
func TestCalculateUiPrefixLength_Unknown(t *testing.T) {
	ui := &testProgressBarUi{}

	expected := 0
	if res := calculateUiPrefixLength(ui); res != expected {
		t.Fatalf("calculateUiPrefixLength should have returned a length of %d", expected)
	}
}

func TestCalculateUiPrefixLength_BasicUi(t *testing.T) {
	ui := &packer.BasicUi{}

	expected := 1
	if res := calculateUiPrefixLength(ui); res != expected {
		t.Fatalf("calculateUiPrefixLength should have returned a length of %d", expected)
	}
}

func TestCalculateUiPrefixLength_TargetedUI(t *testing.T) {
	ui := &packer.TargetedUI{}
	ui.Target = "TestTarget"
	arrowText := "==>"

	expected := len(arrowText + " " + ui.Target + ": ")
	if res := calculateUiPrefixLength(ui); res != expected {
		t.Fatalf("calculateUiPrefixLength should have returned a length of %d", expected)
	}
}

func TestCalculateUiPrefixLength_TargetedUIWrappingBasicUi(t *testing.T) {
	ui := &packer.TargetedUI{}
	ui.Target = "TestTarget"
	ui.Ui = &packer.BasicUi{}
	arrowText := "==>"

	expected := len(arrowText + " " + ui.Target + ": " + "\n")
	if res := calculateUiPrefixLength(ui); res != expected {
		t.Fatalf("calculateUiPrefixLength should have returned a length of %d", expected)
	}
}

func TestCalculateUiPrefixLength_TargetedUIWrappingMachineUi(t *testing.T) {
	ui := &packer.TargetedUI{}
	ui.Target = "TestTarget"
	ui.Ui = &packer.MachineReadableUi{}

	expected := 0
	if res := calculateUiPrefixLength(ui); res != expected {
		t.Fatalf("calculateUiPrefixLength should have returned a length of %d", expected)
	}
}
func TestDefaultProgressBar(t *testing.T) {
	var callbackCalled bool

	// Initialize the default progress bar
	bar := GetDefaultProgressBar()
	bar.Callback = func(state string) {
		callbackCalled = true
		t.Logf("TestDefaultProgressBar emitted %#v", state)
	}
	bar.SetTotal64(1)

	// Set it off
	progressBar := bar.Start()
	progressBar.Set64(1)

	// Check to see that the callback was hit
	if !callbackCalled {
		t.Fatalf("TestDefaultProgressBar.Callback should be called")
	}
}

func TestDummyProgressBar(t *testing.T) {
	var callbackCalled bool

	// Initialize the dummy progress bar
	bar := GetDummyProgressBar()
	bar.Callback = func(state string) {
		callbackCalled = true
		t.Logf("TestDummyProgressBar emitted %#v", state)
	}
	bar.SetTotal64(1)

	// Now we can go
	progressBar := bar.Start()
	progressBar.Set64(1)

	// Check to see that the callback was hit
	if callbackCalled {
		t.Fatalf("TestDummyProgressBar.Callback should not be called")
	}
}

func TestUiProgressBar(t *testing.T) {

	ui := &testProgressBarUi{}

	// Initialize the Ui progress bar
	bar := GetProgressBar(ui, nil)
	bar.SetTotal64(1)

	// Ensure that callback has been set to something
	if bar.Callback == nil {
		t.Fatalf("TestUiProgressBar.Callback should be initialized")
	}

	// Now we can go
	progressBar := bar.Start()
	progressBar.Set64(1)

	// Check to see that the callback was hit
	if !ui.messageCalled {
		t.Fatalf("TestUiProgressBar.messageCalled should be called")
	}
	t.Logf("TestUiProgressBar emitted %#v", ui.messageMessage)
}
