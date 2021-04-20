package ncloud

import (
	"context"
	"fmt"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type LoginKey struct {
	KeyName    string
	PrivateKey string
}

type StepCreateLoginKey struct {
	Conn           *NcloudAPIClient
	CreateLoginKey func() (*LoginKey, error)
	DeleteLoginKey func(loginKey string) error
	Say            func(message string)
	Error          func(e error)
	Config         *Config
}

func NewStepCreateLoginKey(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepCreateLoginKey {
	var step = &StepCreateLoginKey{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.CreateLoginKey = step.createVpcLoginKey
		step.DeleteLoginKey = step.deleteVpcLoginKey
	} else {
		step.CreateLoginKey = step.createClassicLoginKey
		step.DeleteLoginKey = step.deleteClassicLoginKey
	}

	return step
}

func (s *StepCreateLoginKey) createClassicLoginKey() (*LoginKey, error) {
	keyName := fmt.Sprintf("packer-%d", time.Now().Unix())
	reqParams := &server.CreateLoginKeyRequest{KeyName: &keyName}

	privateKey, err := s.Conn.server.V2Api.CreateLoginKey(reqParams)
	if err != nil {
		return nil, err
	}

	return &LoginKey{keyName, *privateKey.PrivateKey}, nil
}

func (s *StepCreateLoginKey) createVpcLoginKey() (*LoginKey, error) {
	keyName := fmt.Sprintf("packer-%d", time.Now().Unix())
	reqParams := &vserver.CreateLoginKeyRequest{KeyName: &keyName}

	privateKey, err := s.Conn.vserver.V2Api.CreateLoginKey(reqParams)
	if err != nil {
		return nil, err
	}

	return &LoginKey{keyName, *privateKey.PrivateKey}, nil
}

func (s *StepCreateLoginKey) deleteClassicLoginKey(keyName string) error {
	reqParams := &server.DeleteLoginKeyRequest{KeyName: &keyName}
	_, err := s.Conn.server.V2Api.DeleteLoginKey(reqParams)

	if err != nil {
		return err
	}

	return nil
}

func (s *StepCreateLoginKey) deleteVpcLoginKey(keyName string) error {
	reqParams := &vserver.DeleteLoginKeysRequest{KeyNameList: []*string{&keyName}}
	_, err := s.Conn.vserver.V2Api.DeleteLoginKeys(reqParams)

	if err != nil {
		return err
	}

	return nil
}

func (s *StepCreateLoginKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create Login Key")

	loginKey, err := s.CreateLoginKey()
	if err == nil {
		state.Put("login_key", loginKey)
		s.Say(fmt.Sprintf("Login Key[%s] is created", loginKey.KeyName))
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreateLoginKey) Cleanup(state multistep.StateBag) {
	if loginKey, ok := state.GetOk("login_key"); ok {
		s.Say("Clean up login key")
		if err := s.DeleteLoginKey(loginKey.(*LoginKey).KeyName); err != nil {
			s.Error(err)
			return
		}
	}
}
