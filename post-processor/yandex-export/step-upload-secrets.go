package yandexexport

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/yandex"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
)

type StepUploadSecrets struct{}

const (
	sharedAWSCredFile = "/tmp/aws-credentials"
)

// Run reads the instance metadata and looks for the log entry
// indicating the cloud-init script finished.
func (s *StepUploadSecrets) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	_ = state.Get("config").(*yandex.Config)
	_ = state.Get("driver").(yandex.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	comm := state.Get("communicator").(packersdk.Communicator)
	s3Secret := state.Get("s3_secret").(*awscompatibility.CreateAccessKeyResponse)

	ui.Say("Upload secrets..")
	creds := fmt.Sprintf(
		"[default]\naws_access_key_id = %s\naws_secret_access_key = %s\n",
		s3Secret.GetAccessKey().GetKeyId(),
		s3Secret.GetSecret())

	err := comm.Upload(sharedAWSCredFile, strings.NewReader(creds), nil)
	if err != nil {
		return yandex.StepHaltWithError(state, err)
	}
	ui.Message("Secrets has been uploaded")

	return multistep.ActionContinue
}

// Cleanup.
func (s *StepUploadSecrets) Cleanup(state multistep.StateBag) {}
