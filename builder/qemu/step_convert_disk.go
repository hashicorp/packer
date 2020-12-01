package qemu

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/retry"

	"os"
)

// This step converts the virtual disk that was used as the
// hard drive for the virtual machine.
type stepConvertDisk struct {
	DiskCompression bool
	Format          string
	OutputDir       string
	SkipCompaction  bool
	VMName          string

	QemuImgArgs QemuImgArgs
}

func (s *stepConvertDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	diskName := s.VMName

	if s.SkipCompaction && !s.DiskCompression {
		return multistep.ActionContinue
	}

	name := diskName + ".convert"

	sourcePath := filepath.Join(s.OutputDir, diskName)
	targetPath := filepath.Join(s.OutputDir, name)

	command := s.buildConvertCommand(sourcePath, targetPath)

	ui.Say("Converting hard drive...")
	// Retry the conversion a few times in case it takes the qemu process a
	// moment to release the lock
	err := retry.Config{
		Tries: 10,
		ShouldRetry: func(err error) bool {
			if strings.Contains(err.Error(), `Failed to get shared "write" lock`) {
				ui.Say("Error getting file lock for conversion; retrying...")
				return true
			}
			return false
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 1 * time.Second, MaxBackoff: 10 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		return driver.QemuImg(command...)
	})

	if err != nil {
		switch err.(type) {
		case *retry.RetryExhaustedError:
			err = fmt.Errorf("Exhausted retries for getting file lock: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		default:
			err := fmt.Errorf("Error converting hard drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if err := os.Rename(targetPath, sourcePath); err != nil {
		err := fmt.Errorf("Error moving converted hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepConvertDisk) buildConvertCommand(sourcePath, targetPath string) []string {
	command := []string{"convert"}

	if s.DiskCompression {
		command = append(command, "-c")
	}

	// Add user-provided convert args
	command = append(command, s.QemuImgArgs.Convert...)

	// Add format, and paths.
	command = append(command, "-O", s.Format, sourcePath, targetPath)

	return command
}

func (s *stepConvertDisk) Cleanup(state multistep.StateBag) {}
