package jdcloud

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
)

type stepConfigCredentials struct {
	InstanceSpecConfig *JDCloudInstanceSpecConfig
	ui                 packer.Ui
}

func (s *stepConfigCredentials) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	s.ui = state.Get("ui").(packer.Ui)
	password := s.InstanceSpecConfig.Comm.SSHPassword
	privateKeyPath := s.InstanceSpecConfig.Comm.SSHPrivateKeyFile
	privateKeyName := s.InstanceSpecConfig.Comm.SSHKeyPairName
	newKeyName := s.InstanceSpecConfig.Comm.SSHTemporaryKeyPairName

	if len(privateKeyPath) > 0 && len(privateKeyName) > 0 {
		s.ui.Message("\t Private key detected, we are going to login with this private key :)")
		return s.ReadExistingPair()
	}

	if len(newKeyName) > 0 {
		s.ui.Message("\t We are going to create a new key pair with its name=" + newKeyName)
		return s.CreateRandomKeyPair(newKeyName)
	}

	if len(password) > 0 {
		s.ui.Message("\t Password detected, we are going to login with this password :)")
		return multistep.ActionContinue
	}

	s.ui.Error("[ERROR] Didn't detect any credentials, you have to specify either " +
		"{password} or " +
		"{key_name+local_private_key_path} or " +
		"{temporary_key_pair_name} cheers :)")
	return multistep.ActionHalt
}

func (s *stepConfigCredentials) ReadExistingPair() multistep.StepAction {
	privateKeyBytes, err := ioutil.ReadFile(s.InstanceSpecConfig.Comm.SSHPrivateKeyFile)
	if err != nil {
		s.ui.Error("Cannot read local private-key, were they correctly placed? Here's the error" + err.Error())
		return multistep.ActionHalt
	}
	s.ui.Message("\t\t Keys read successfully :)")
	s.InstanceSpecConfig.Comm.SSHPrivateKey = privateKeyBytes
	return multistep.ActionContinue
}

func (s *stepConfigCredentials) CreateRandomKeyPair(keyName string) multistep.StepAction {
	req := apis.NewCreateKeypairRequest(Region, keyName)
	resp, err := VmClient.CreateKeypair(req)
	if err != nil || resp.Error.Code != FINE {
		s.ui.Error(fmt.Sprintf("[ERROR] Cannot create a new key pair for you, \n error=%v \n response=%v", err, resp))
		return multistep.ActionHalt
	}
	s.ui.Message("\t\t Keys created successfully :)")
	s.InstanceSpecConfig.Comm.SSHPrivateKey = []byte(resp.Result.PrivateKey)
	s.InstanceSpecConfig.Comm.SSHKeyPairName = keyName
	return multistep.ActionContinue
}

func (s *stepConfigCredentials) Cleanup(state multistep.StateBag) {

}
