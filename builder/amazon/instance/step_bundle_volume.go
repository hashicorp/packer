package instance

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type bundleCmdData struct {
	AccountId    string
	Architecture string
	CertPath     string
	Destination  string
	KeyPath      string
	Prefix       string
	PrivatePath  string
}

type StepBundleVolume struct{}

func (s *StepBundleVolume) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	config := state["config"].(*Config)
	instance := state["instance"].(*ec2.Instance)
	ui := state["ui"].(packer.Ui)
	x509RemoteCertPath := state["x509RemoteCertPath"].(string)
	x509RemoteKeyPath := state["x509RemoteKeyPath"].(string)

	// Bundle the volume
	var err error
	config.BundleVolCommand, err = config.tpl.Process(config.BundleVolCommand, bundleCmdData{
		AccountId:    config.AccountId,
		Architecture: instance.Architecture,
		CertPath:     x509RemoteCertPath,
		Destination:  config.BundleDestination,
		KeyPath:      x509RemoteKeyPath,
		Prefix:       config.BundlePrefix,
		PrivatePath:  config.X509UploadPath,
	})
	if err != nil {
		err := fmt.Errorf("Error processing bundle volume command: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Bundling the volume...")
	cmd := new(packer.RemoteCmd)
	cmd.Command = config.BundleVolCommand
	if err := cmd.StartWithUi(comm, ui); err != nil {
		state["error"] = fmt.Errorf("Error bundling volume: %s", err)
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	if cmd.ExitStatus != 0 {
		state["error"] = fmt.Errorf(
			"Volume bundling failed. Please see the output above for more\n" +
				"details on what went wrong.")
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	// Store the manifest path
	manifestName := config.BundlePrefix + ".manifest.xml"
	state["manifest_name"] = manifestName
	state["manifest_path"] = fmt.Sprintf(
		"%s/%s", config.BundleDestination, manifestName)

	return multistep.ActionContinue
}

func (s *StepBundleVolume) Cleanup(map[string]interface{}) {}
