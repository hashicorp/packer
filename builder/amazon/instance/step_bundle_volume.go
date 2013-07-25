package instance

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"text/template"
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

	// Verify the AMI tools are available
	ui.Say("Checking for EC2 AMI tools...")
	cmd := &packer.RemoteCmd{Command: "ec2-ami-tools-version"}
	if err := comm.Start(cmd); err != nil {
		state["error"] = fmt.Errorf("Error checking for AMI tools: %s", err)
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}
	cmd.Wait()

	if cmd.ExitStatus != 0 {
		state["error"] = fmt.Errorf(
			"The EC2 AMI tools could not be detected. These must be manually\n" +
				"via a provisioner or some other means and are required for Packer\n" +
				"to create an instance-store AMI.")
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	// Bundle the volume
	var bundleCmd bytes.Buffer
	tData := bundleCmdData{
		AccountId:    config.AccountId,
		Architecture: instance.Architecture,
		CertPath:     x509RemoteCertPath,
		Destination:  config.BundleDestination,
		KeyPath:      x509RemoteKeyPath,
		Prefix:       config.BundlePrefix,
		PrivatePath:  config.X509UploadPath,
	}
	t := template.Must(template.New("bundleCmd").Parse(config.BundleVolCommand))
	t.Execute(&bundleCmd, tData)

	ui.Say("Bundling the volume...")
	cmd = new(packer.RemoteCmd)
	cmd.Command = bundleCmd.String()
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

	return multistep.ActionContinue
}

func (s *StepBundleVolume) Cleanup(map[string]interface{}) {}
