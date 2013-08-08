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
	SecretKey       string
}

type StepUploadBundle struct{}

func (s *StepUploadBundle) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	config := state["config"].(*Config)
	manifestName := state["manifest_name"].(string)
	manifestPath := state["manifest_path"].(string)
	ui := state["ui"].(packer.Ui)

	var err error
	config.BundleUploadCommand, err = config.tpl.Process(config.BundleUploadCommand, uploadCmdData{
		AccessKey:       config.AccessKey,
		BucketName:      config.S3Bucket,
		BundleDirectory: config.BundleDestination,
		ManifestPath:    manifestPath,
		SecretKey:       config.SecretKey,
	})
	if err != nil {
		err := fmt.Errorf("Error processing bundle upload command: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Uploading the bundle...")
	cmd := &packer.RemoteCmd{Command: config.BundleUploadCommand}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		state["error"] = fmt.Errorf("Error uploading volume: %s", err)
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	if cmd.ExitStatus != 0 {
		state["error"] = fmt.Errorf(
			"Bundle upload failed. Please see the output above for more\n" +
				"details on what went wrong.")
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	state["remote_manifest_path"] = fmt.Sprintf(
		"%s/%s", config.S3Bucket, manifestName)

	return multistep.ActionContinue
}

func (s *StepUploadBundle) Cleanup(state map[string]interface{}) {}
