package instance

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/multistep"
)

type uploadCmdData struct {
	AccessKey       string
	BucketName      string
	BundleDirectory string
	ManifestPath    string
	Region          string
	SecretKey       string
}

type StepUploadBundle struct {
	Debug bool
}

func (s *StepUploadBundle) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	config := state.Get("config").(*Config)
	manifestName := state.Get("manifest_name").(string)
	manifestPath := state.Get("manifest_path").(string)
	ui := state.Get("ui").(packer.Ui)

	region, err := config.Region()
	if err != nil {
		err := fmt.Errorf("Error retrieving region: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	accessKey := config.AccessKey
	secretKey := config.SecretKey
	session, err := config.AccessConfig.Session()
	accessConfig := session.Config
	if err == nil && accessKey == "" && secretKey == "" {
		credentials, err := accessConfig.Credentials.Get()
		if err == nil {
			accessKey = credentials.AccessKeyID
			secretKey = credentials.SecretAccessKey
		}
	}

	config.ctx.Data = uploadCmdData{
		AccessKey:       accessKey,
		BucketName:      config.S3Bucket,
		BundleDirectory: config.BundleDestination,
		ManifestPath:    manifestPath,
		Region:          region,
		SecretKey:       secretKey,
	}
	config.BundleUploadCommand, err = interpolate.Render(config.BundleUploadCommand, &config.ctx)
	if err != nil {
		err := fmt.Errorf("Error processing bundle upload command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Uploading the bundle...")
	cmd := &packer.RemoteCmd{Command: config.BundleUploadCommand}

	if s.Debug {
		ui.Say(fmt.Sprintf("Running: %s", config.BundleUploadCommand))
	}

	if err := cmd.StartWithUi(comm, ui); err != nil {
		state.Put("error", fmt.Errorf("Error uploading volume: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	if cmd.ExitStatus != 0 {
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
