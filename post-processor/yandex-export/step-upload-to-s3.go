package yandexexport

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/yandex"
)

type StepUploadToS3 struct {
	Paths []string
}

// Run reads the instance metadata and looks for the log entry
// indicating the cloud-init script finished.
func (s *StepUploadToS3) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	comm := state.Get("communicator").(packersdk.Communicator)

	cmdUploadToS3 := &packersdk.RemoteCmd{
		Command: fmt.Sprintf(
			"%s=%s aws s3 --region=%s --endpoint-url=https://%s cp disk.qcow2 %s",
			"AWS_SHARED_CREDENTIALS_FILE",
			sharedAWSCredFile,
			defaultStorageRegion,
			defaultStorageEndpoint,
			s.Paths[0],
		),
	}
	ui.Say("Upload to S3...")
	if err := cmdUploadToS3.RunWithUi(ctx, comm, ui); err != nil {
		return yandex.StepHaltWithError(state, err)
	}
	if cmdUploadToS3.ExitStatus() != 0 {
		return yandex.StepHaltWithError(state, fmt.Errorf("Cannout upload to S3, exit code %d", cmdUploadToS3.ExitStatus()))
	}

	versionExtraFlags, err := getVersionExtraFlags(ctx, comm)
	if err != nil {
		ui.Message(fmt.Sprintf("[WARN] Cannot upload to other storage: %s", err))
		return multistep.ActionContinue
	}
	wg := new(sync.WaitGroup)
	defer wg.Wait()
	for _, path := range s.Paths[1:] {

		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			ui.Message(fmt.Sprintf("Start copy %s to %s...", s.Paths[0], path))
			cmd := &packersdk.RemoteCmd{
				Command: fmt.Sprintf(
					"%s=%s aws s3 --region=%s --endpoint-url=https://%s cp %s %s %s",
					"AWS_SHARED_CREDENTIALS_FILE",
					sharedAWSCredFile,
					defaultStorageRegion,
					defaultStorageEndpoint,
					versionExtraFlags,
					s.Paths[0],
					path,
				),
			}
			if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
				ui.Message(fmt.Sprintf("[WARN] Failed upload to %s", path))
			}
			if cmd.ExitStatus() != 0 {
				ui.Message(fmt.Sprintf("[WARN] Failed upload to %s", path))
			}
		}(path)
	}

	return multistep.ActionContinue
}

// Cleanup nothing
func (s *StepUploadToS3) Cleanup(state multistep.StateBag) {}

func getVersionExtraFlags(ctx context.Context, comm packersdk.Communicator) (string, error) {
	buff := new(bytes.Buffer)
	cmd := &packersdk.RemoteCmd{
		Command: "aws --version",
		Stdout:  buff,
	}
	if err := comm.Start(ctx, cmd); err != nil {
		return "", err
	}
	if cmd.Wait() != 0 {
		return "", fmt.Errorf("Cannot detect aws version")
	}
	vsn := buff.String()
	switch {
	case strings.HasPrefix(vsn, "aws-cli/2."):
		return "--copy-props metadata-directive", nil
	}
	return "", nil
}
