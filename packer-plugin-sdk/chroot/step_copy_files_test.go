package chroot

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// testUI returns a test ui plus a function to retrieve the errors written to the ui
func testUI() (packersdk.Ui, func() string) {
	errorBuffer := &strings.Builder{}
	ui := &packer.BasicUi{
		Reader:      strings.NewReader(""),
		Writer:      ioutil.Discard,
		ErrorWriter: errorBuffer,
	}
	return ui, errorBuffer.String
}

func TestCopyFilesCleanupFunc_ImplementsCleanupFunc(t *testing.T) {
	var raw interface{}
	raw = new(StepCopyFiles)
	if _, ok := raw.(Cleanup); !ok {
		t.Fatalf("cleanup func should be a CleanupFunc")
	}
}

func TestCopyFiles_Run(t *testing.T) {
	mountPath := "/mnt/abcde"
	copySource := "/etc/resolv.conf"
	copyDestination := path.Join(mountPath, "etc", "resolv.conf")

	step := &StepCopyFiles{
		Files: []string{
			copySource,
		},
	}

	var gotCommand string
	commandRunCount := 0
	var wrapper common.CommandWrapper
	wrapper = func(ran string) (string, error) {
		gotCommand = ran
		commandRunCount++
		return "", nil
	}

	state := new(multistep.BasicStateBag)
	state.Put("mount_path", mountPath)
	state.Put("wrappedCommand", wrapper)

	ui, getErrs := testUI()
	state.Put("ui", ui)

	var expectedCopyTemplate string

	switch runtime.GOOS {
	case "linux":
		expectedCopyTemplate = "cp --remove-destination %s %s"
	case "freebsd":
		expectedCopyTemplate = "cp -f %s %s"
	default:
		t.Skip("Unsupported operating system")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got := step.Run(ctx, state)
	if got != multistep.ActionContinue {
		t.Errorf("Expected 'continue', but got '%v'", got)
	}

	if commandRunCount != 1 {
		t.Errorf("Copy command should run exactly once but ran %v times", commandRunCount)
	}

	expectedCopyCommand := fmt.Sprintf(expectedCopyTemplate, copySource, copyDestination)
	if gotCommand != expectedCopyCommand {
		t.Errorf("Expected command was '%v' but actual was '%v'", expectedCopyCommand, gotCommand)
	}

	_ = getErrs
}

func TestCopyFiles_CopyNothing(t *testing.T) {
	step := &StepCopyFiles{
		Files: []string{},
	}

	commandRunCount := 0
	var wrapper common.CommandWrapper
	wrapper = func(ran string) (string, error) {
		commandRunCount++
		return "", nil
	}

	state := new(multistep.BasicStateBag)
	state.Put("mount_path", "/mnt/something")
	state.Put("wrappedCommand", wrapper)

	ui, getErrs := testUI()
	state.Put("ui", ui)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got := step.Run(ctx, state)
	if got != multistep.ActionContinue {
		t.Errorf("Expected 'continue', but got '%v'", got)
	}

	if commandRunCount != 0 {
		t.Errorf("Copy command should not run but ran %v times", commandRunCount)
	}

	_ = getErrs
}
