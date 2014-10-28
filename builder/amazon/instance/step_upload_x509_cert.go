package instance

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
)

type StepUploadX509Cert struct{}

func (s *StepUploadX509Cert) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	x509RemoteCertPath := config.X509UploadPath + "/cert.pem"
	x509RemoteKeyPath := config.X509UploadPath + "/key.pem"

	ui.Say("Uploading X509 Certificate...")
	if err := s.uploadSingle(comm, x509RemoteCertPath, config.X509CertPath); err != nil {
		state.Put("error", fmt.Errorf("Error uploading X509 cert: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	if err := s.uploadSingle(comm, x509RemoteKeyPath, config.X509KeyPath); err != nil {
		state.Put("error", fmt.Errorf("Error uploading X509 cert: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	state.Put("x509RemoteCertPath", x509RemoteCertPath)
	state.Put("x509RemoteKeyPath", x509RemoteKeyPath)

	return multistep.ActionContinue
}

func (s *StepUploadX509Cert) Cleanup(multistep.StateBag) {}

func (s *StepUploadX509Cert) uploadSingle(comm packer.Communicator, dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return comm.Upload(dst, f, nil)
}
