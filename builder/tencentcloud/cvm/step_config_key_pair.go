package cvm

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepConfigKeyPair struct {
	Debug        bool
	Comm         *communicator.Config
	DebugKeyPath string
	keyID        string
}

func (s *stepConfigKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("cvm_client").(*cvm.Client)

	if s.Comm.SSHPrivateKeyFile != "" {
		Say(state, "Using existing SSH private key", "")
		privateKeyBytes, err := ioutil.ReadFile(s.Comm.SSHPrivateKeyFile)
		if err != nil {
			return Halt(state, err, fmt.Sprintf("Failed to load configured private key(%s)", s.Comm.SSHPrivateKeyFile))
		}
		s.Comm.SSHPrivateKey = privateKeyBytes
		Message(state, fmt.Sprintf("Loaded %d bytes private key data", len(s.Comm.SSHPrivateKey)), "")
		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth {
		if s.Comm.SSHKeyPairName == "" {
			Say(state, "Using SSH agent with key pair in source image", "")
			return multistep.ActionContinue
		}
		Say(state, fmt.Sprintf("Using SSH agent with exists key pair(%s)", s.Comm.SSHKeyPairName), "")
		return multistep.ActionContinue
	}

	if s.Comm.SSHTemporaryKeyPairName == "" {
		Say(state, "Not to use temporary keypair", "")
		s.Comm.SSHKeyPairName = ""
		return multistep.ActionContinue
	}

	Say(state, s.Comm.SSHTemporaryKeyPairName, "Trying to create a new keypair")

	req := cvm.NewCreateKeyPairRequest()
	req.KeyName = &s.Comm.SSHTemporaryKeyPairName
	defaultProjectId := int64(0)
	req.ProjectId = &defaultProjectId
	var resp *cvm.CreateKeyPairResponse
	err := Retry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = client.CreateKeyPair(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to create keypair")
	}

	// set keyId to delete when Cleanup
	s.keyID = *resp.Response.KeyPair.KeyId
	state.Put("temporary_key_pair_id", s.keyID)
	Message(state, s.keyID, "Keypair created")

	s.Comm.SSHKeyPairName = *resp.Response.KeyPair.KeyId
	s.Comm.SSHPrivateKey = []byte(*resp.Response.KeyPair.PrivateKey)

	if s.Debug {
		Message(state, fmt.Sprintf("Saving temporary key to %s for debug purposes", s.DebugKeyPath), "")
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			return Halt(state, err, "Failed to saving debug key file")
		}
		defer f.Close()
		if _, err := f.Write([]byte(*resp.Response.KeyPair.PrivateKey)); err != nil {
			return Halt(state, err, "Failed to writing debug key file")
		}
		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				return Halt(state, err, "Failed to chmod debug key file")
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

	SayClean(state, "keypair")

	req := cvm.NewDeleteKeyPairsRequest()
	req.KeyIds = []*string{&s.keyID}
	err := Retry(ctx, func(ctx context.Context) error {
		_, e := client.DeleteKeyPairs(req)
		return e
	})
	if err != nil {
		Error(state, err, fmt.Sprintf("Failed to delete keypair(%s), please delete it manually", s.keyID))
	}

	if s.Debug {
		if err := os.Remove(s.DebugKeyPath); err != nil {
			Error(state, err, fmt.Sprintf("Failed to delete debug key file(%s), please delete it manually", s.DebugKeyPath))
		}
	}
}
