package qemu

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"os"
)

// This step converts the virtual disk that was used as the
// hard drive for the virtual machine.
type stepConvertDisk struct{}

func (s *stepConvertDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	diskName := state.Get("disk_filename").(string)
	ui := state.Get("ui").(packer.Ui)

	if config.SkipCompaction && !config.DiskCompression {
		return multistep.ActionContinue
	}

	name := diskName + ".convert"

	sourcePath := filepath.Join(config.OutputDir, diskName)
	targetPath := filepath.Join(config.OutputDir, name)

	command := []string{
		"convert",
	}

	if config.DiskCompression {
		command = append(command, "-c")
	}

	command = append(command, []string{
		"-O", config.Format,
		sourcePath,
		targetPath,
	}...,
	)

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
		if err == common.RetryExhaustedError {
			err = fmt.Errorf("Exhausted retries for getting file lock: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else {
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

func (s *stepConvertDisk) Cleanup(state multistep.StateBag) {}
