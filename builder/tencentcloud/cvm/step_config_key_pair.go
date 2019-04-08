package cvm

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepConfigKeyPair struct {
	Debug        bool
	Comm         *communicator.Config
	DebugKeyPath string

	keyID string
}

func (s *stepConfigKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if s.Comm.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
		privateKeyBytes, err := ioutil.ReadFile(s.Comm.SSHPrivateKeyFile)
		if err != nil {
			state.Put("error", fmt.Errorf(
				"loading configured private key file failed: %s", err))
			return multistep.ActionHalt
		}

		s.Comm.SSHPrivateKey = privateKeyBytes

		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth && s.Comm.SSHKeyPairName == "" {
		ui.Say("Using SSH Agent with key pair in source image")
		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth && s.Comm.SSHKeyPairName != "" {
		ui.Say(fmt.Sprintf("Using SSH Agent for existing key pair %s", s.Comm.SSHKeyPairName))
		return multistep.ActionContinue
	}

	if s.Comm.SSHTemporaryKeyPairName == "" {
		ui.Say("Not using temporary keypair")
		s.Comm.SSHKeyPairName = ""
		return multistep.ActionContinue
	}

	client := state.Get("cvm_client").(*cvm.Client)
	ui.Say(fmt.Sprintf("Creating temporary keypair: %s", s.Comm.SSHTemporaryKeyPairName))
	req := cvm.NewCreateKeyPairRequest()
	req.KeyName = &s.Comm.SSHTemporaryKeyPairName
	defaultProjectId := int64(0)
	req.ProjectId = &defaultProjectId
	resp, err := client.CreateKeyPair(req)
	if err != nil {
		state.Put("error", fmt.Errorf("creating temporary keypair failed: %s", err.Error()))
		return multistep.ActionHalt
	}

	// set keyId to delete when Cleanup
	s.keyID = *resp.Response.KeyPair.KeyId

	s.Comm.SSHKeyPairName = *resp.Response.KeyPair.KeyId
	s.Comm.SSHPrivateKey = []byte(*resp.Response.KeyPair.PrivateKey)

	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("creating debug key file failed:%s", err.Error()))
			return multistep.ActionHalt
		}
		defer f.Close()

		if _, err := f.Write([]byte(*resp.Response.KeyPair.PrivateKey)); err != nil {
			state.Put("error", fmt.Errorf("writing debug key file failed:%s", err.Error()))
			return multistep.ActionHalt
		}

		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				state.Put("error", fmt.Errorf("setting debug key file's permission failed:%s", err.Error()))
				return multistep.ActionHalt
			}
		}
	}
	return multistep.ActionContinue
}

func (s *stepConfigKeyPair) Cleanup(state multistep.StateBag) {
	if s.Comm.SSHPrivateKeyFile != "" || (s.Comm.SSHKeyPairName == "" && s.keyID == "") {
		return
	}
	ctx := context.TODO()

	client := state.Get("cvm_client").(*cvm.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary keypair...")
	req := cvm.NewDeleteKeyPairsRequest()
	req.KeyIds = []*string{&s.keyID}
	err := retry.Config{
		Tries:      60,
		RetryDelay: (&retry.Backoff{InitialBackoff: 5 * time.Second, MaxBackoff: 5 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		_, err := client.DeleteKeyPairs(req)
		return err
	})
	if err != nil {
		ui.Error(fmt.Sprintf(
			"delete keypair failed, please delete it manually, keyId: %s, err: %s", s.keyID, err.Error()))
	}
	if s.Debug {
		if err := os.Remove(s.DebugKeyPath); err != nil {
			ui.Error(fmt.Sprintf("delete debug key file %s failed: %s", s.DebugKeyPath, err.Error()))
		}
	}
}
