package instance

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
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

type StepBundleVolume struct {
	Debug bool
}

func (s *StepBundleVolume) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	config := state.Get("config").(*Config)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)
	x509RemoteCertPath := state.Get("x509RemoteCertPath").(string)
	x509RemoteKeyPath := state.Get("x509RemoteKeyPath").(string)

	// Bundle the volume
	var err error
	config.ctx.Data = bundleCmdData{
		AccountId:    config.AccountId,
		Architecture: *instance.Architecture,
		CertPath:     x509RemoteCertPath,
		Destination:  config.BundleDestination,
		KeyPath:      x509RemoteKeyPath,
		Prefix:       config.BundlePrefix,
		PrivatePath:  config.X509UploadPath,
	}
	config.BundleVolCommand, err = interpolate.Render(config.BundleVolCommand, &config.ctx)
	if err != nil {
		err := fmt.Errorf("Error processing bundle volume command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Bundling the volume...")
	cmd := new(packer.RemoteCmd)
	cmd.Command = config.BundleVolCommand

	if s.Debug {
		ui.Say(fmt.Sprintf("Running: %s", config.BundleVolCommand))
	}

	if err := cmd.StartWithUi(comm, ui); err != nil {
		state.Put("error", fmt.Errorf("Error bundling volume: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	if cmd.ExitStatus != 0 {
		state.Put("error", fmt.Errorf(
			"Volume bundling failed. Please see the output above for more\n"+
				"details on what went wrong.\n\n"+
				"One common cause for this error is ec2-bundle-vol not being\n"+
				"available on the target instance."))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	// Store the manifest path
	manifestName := config.BundlePrefix + ".manifest.xml"
	state.Put("manifest_name", manifestName)
	state.Put("manifest_path", fmt.Sprintf(
		"%s/%s", config.BundleDestination, manifestName))

	return multistep.ActionContinue
}

func (s *StepBundleVolume) Cleanup(multistep.StateBag) {}
