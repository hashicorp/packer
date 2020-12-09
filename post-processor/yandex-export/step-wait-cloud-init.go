package yandexexport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer/builder/yandex"
)

type StepWaitCloudInitScript int

type cloudInitStatus struct {
	V1 struct {
		Errors []interface{}
	}
}

type cloudInitError struct {
	Err error
}

func (e *cloudInitError) Error() string {
	return e.Err.Error()
}

// Run reads the instance metadata and looks for the log entry
// indicating the cloud-init script finished.
func (s *StepWaitCloudInitScript) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	comm := state.Get("communicator").(packersdk.Communicator)

	ui.Say("Waiting for any running cloud-init script to finish...")

	ctxWithCancel, cancelCtx := context.WithCancel(ctx)

	defer cancelCtx()

	go func() {
		cmd := &packersdk.RemoteCmd{
			Command: "tail -f /var/log/cloud-init-output.log",
		}

		err := cmd.RunWithUi(ctxWithCancel, comm, ui)
		if err != nil && !errors.Is(err, context.Canceled) {
			ui.Error(err.Error())
			return
		}
		ui.Message("Init output closed")
	}()

	// Keep checking the serial port output to see if the cloud-init script is done.
	retryConfig := &retry.Config{
		ShouldRetry: func(e error) bool {
			switch e.(type) {
			case *cloudInitError:
				return false
			}
			return true
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 10 * time.Second, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}

	err := retryConfig.Run(ctx, func(ctx context.Context) error {
		buff := bytes.Buffer{}
		err := comm.Download("/var/run/cloud-init/result.json", &buff)
		if err != nil {
			err := fmt.Errorf("Waiting cloud-init script status: %s", err)
			return err
		}
		result := &cloudInitStatus{}
		err = json.Unmarshal(buff.Bytes(), result)
		if err != nil {
			err := fmt.Errorf("Failed parse result: %s", err)
			return &cloudInitError{Err: err}
		}
		if len(result.V1.Errors) != 0 {
			err := fmt.Errorf("Result: %v", result.V1.Errors)
			return &cloudInitError{Err: err}
		}
		return nil
	})

	if err != nil {
		err := fmt.Errorf("Error waiting for cloud-init script to finish: %s", err)
		return yandex.StepHaltWithError(state, err)
	}
	ui.Say("Cloud-init script has finished running.")
	return multistep.ActionContinue
}

// Cleanup.
func (s *StepWaitCloudInitScript) Cleanup(state multistep.StateBag) {}
