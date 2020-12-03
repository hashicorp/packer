package ncloud

import (
	"context"
	"fmt"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type LoginKey struct {
	KeyName    string
	PrivateKey string
}

type StepCreateLoginKey struct {
	Conn           *NcloudAPIClient
	CreateLoginKey func() (*LoginKey, error)
	Say            func(message string)
	Error          func(e error)
}

func NewStepCreateLoginKey(conn *NcloudAPIClient, ui packersdk.Ui) *StepCreateLoginKey {
	var step = &StepCreateLoginKey{
		Conn:  conn,
		Say:   func(message string) { ui.Say(message) },
		Error: func(e error) { ui.Error(e.Error()) },
	}

	step.CreateLoginKey = step.createLoginKey

	return step
}

func (s *StepCreateLoginKey) createLoginKey() (*LoginKey, error) {
	keyName := fmt.Sprintf("packer-%d", time.Now().Unix())
	reqParams := &server.CreateLoginKeyRequest{KeyName: &keyName}

	privateKey, err := s.Conn.server.V2Api.CreateLoginKey(reqParams)
	if err != nil {
		return nil, err
	}

	return &LoginKey{keyName, *privateKey.PrivateKey}, nil
}

func (s *StepCreateLoginKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create Login Key")

	loginKey, err := s.CreateLoginKey()
	if err == nil {
		state.Put("LoginKey", loginKey)
		s.Say(fmt.Sprintf("Login Key[%s] is created", loginKey.KeyName))
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreateLoginKey) Cleanup(state multistep.StateBag) {
	if loginKey, ok := state.GetOk("LoginKey"); ok {
		s.Say("Clean up login key")
		reqParams := &server.DeleteLoginKeyRequest{KeyName: &loginKey.(*LoginKey).KeyName}
		s.Conn.server.V2Api.DeleteLoginKey(reqParams)
	}
}
