package instance

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type uploadCmdData struct {
	AccessKey       string
	BucketName      string
	BundleDirectory string
	ManifestPath    string
	S3Endpoint      string
	SecretKey       string
}

type StepUploadBundle struct{}

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

	config.BundleUploadCommand, err = config.tpl.Process(config.BundleUploadCommand, uploadCmdData{
		AccessKey:       config.AccessKey,
		BucketName:      config.S3Bucket,
		BundleDirectory: config.BundleDestination,
		ManifestPath:    manifestPath,
		S3Endpoint:      region.S3Endpoint,
		SecretKey:       config.SecretKey,
	})
	if err != nil {
		err := fmt.Errorf("Error processing bundle upload command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Uploading the bundle...")
	cmd := &packer.RemoteCmd{Command: config.BundleUploadCommand}
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
