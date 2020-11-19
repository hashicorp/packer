package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type uploadCmdData struct {
	AccessKey       string
	BucketName      string
	BundleDirectory string
	ManifestPath    string
	Region          string
	SecretKey       string
	Token           string
}

type StepUploadBundle struct {
	Debug bool
}

func (s *StepUploadBundle) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packersdk.Communicator)
	config := state.Get("config").(*Config)
	manifestName := state.Get("manifest_name").(string)
	manifestPath := state.Get("manifest_path").(string)
	ui := state.Get("ui").(packersdk.Ui)

	accessKey := config.AccessKey
	secretKey := config.SecretKey
	session, err := config.AccessConfig.Session()
	region := *session.Config.Region
	accessConfig := session.Config
	var token string
	if err == nil && accessKey == "" && secretKey == "" {
		credentials, err := accessConfig.Credentials.Get()
		if err == nil {
			accessKey = credentials.AccessKeyID
			secretKey = credentials.SecretAccessKey
			token = credentials.SessionToken
		}
	}

	config.ctx.Data = uploadCmdData{
		AccessKey:       accessKey,
		BucketName:      config.S3Bucket,
		BundleDirectory: config.BundleDestination,
		ManifestPath:    manifestPath,
		Region:          region,
		SecretKey:       secretKey,
		Token:           token,
	}
	config.BundleUploadCommand, err = interpolate.Render(config.BundleUploadCommand, &config.ctx)
	if err != nil {
		err := fmt.Errorf("Error processing bundle upload command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Uploading the bundle...")
	cmd := &packersdk.RemoteCmd{Command: config.BundleUploadCommand}

	if s.Debug {
		ui.Say(fmt.Sprintf("Running: %s", config.BundleUploadCommand))
	}

	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		state.Put("error", fmt.Errorf("Error uploading volume: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	if cmd.ExitStatus() != 0 {
		if cmd.ExitStatus() == 3 {
			ui.Error(fmt.Sprintf("Please check that the bucket `%s` "+
				"does not exist, or exists and is writable. This error "+
				"indicates that the bucket may be owned by somebody else.",
				config.S3Bucket))
		}
		state.Put("error", fmt.Errorf(
			"Bundle upload failed. Please see the output above for more\n"+
				"details on what went wrong."))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	state.Put("remote_manifest_path", fmt.Sprintf(
		"%s/%s", config.S3Bucket, manifestName))

	return multistep.ActionContinue
}

func (s *StepUploadBundle) Cleanup(state multistep.StateBag) {}
