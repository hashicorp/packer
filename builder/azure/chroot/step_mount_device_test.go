package chroot

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepMountDevice_Run(t *testing.T) {
	mountPath, err := ioutil.TempDir("", "stepmountdevicetest")
	if err != nil {
		t.Errorf("Unable to create a temporary directory: %q", err)
	}
	step := &StepMountDevice{
		MountOptions:   []string{"foo"},
		MountPartition: "42",
		MountPath:      mountPath,
	}

	var gotCommand string
	var wrapper common.CommandWrapper
	wrapper = func(ran string) (string, error) {
		gotCommand = ran
		return "", nil
	}

	state := new(multistep.BasicStateBag)
	state.Put("wrappedCommand", wrapper)
	state.Put("device", "/dev/quux")

	ui, getErrs := testUI()
	state.Put("ui", ui)

	var config Config
	state.Put("config", &config)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got := step.Run(ctx, state)
	if got != multistep.ActionContinue {
		t.Errorf("Expected 'continue', but got '%v'", got)
	}

	var expectedMountDevice string
	switch runtime.GOOS {
	case "freebsd":
		expectedMountDevice = "/dev/quuxp42"
	default: // currently just Linux
		expectedMountDevice = "/dev/quux42"
	}
	expectedCommand := fmt.Sprintf("mount -o foo %s %s", expectedMountDevice, mountPath)
	if gotCommand != expectedCommand {
		t.Errorf("Expected '%v', but got '%v'", expectedCommand, gotCommand)
	}

	os.Remove(mountPath)
	_ = getErrs
}
