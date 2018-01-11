package ncloud

import (
	"fmt"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type LoginKey struct {
	KeyName    string
	PrivateKey string
}

type StepCreateLoginKey struct {
	Conn           *ncloud.Conn
	CreateLoginKey func() (*LoginKey, error)
	Say            func(message string)
	Error          func(e error)
}

func NewStepCreateLoginKey(conn *ncloud.Conn, ui packer.Ui) *StepCreateLoginKey {
	var step = &StepCreateLoginKey{
		Conn:  conn,
		Say:   func(message string) { ui.Say(message) },
		Error: func(e error) { ui.Error(e.Error()) },
	}

	step.CreateLoginKey = step.createLoginKey

	return step
}

func (s *StepCreateLoginKey) createLoginKey() (*LoginKey, error) {
	KeyName := fmt.Sprintf("packer-%d", time.Now().Unix())

	privateKey, err := s.Conn.CreateLoginKey(KeyName)
	if err != nil {
		return nil, fmt.Errorf("error code: %d , error message: %s", privateKey.ReturnCode, privateKey.ReturnMessage)
	}

	return &LoginKey{KeyName, privateKey.PrivateKey}, nil
}

func (s *StepCreateLoginKey) Run(state multistep.StateBag) multistep.StepAction {
	s.Say("Create Login Key")

	loginKey, err := s.CreateLoginKey()
	if err == nil {
		state.Put("LoginKey", loginKey)
		s.Say(fmt.Sprintf("Login Key[%s] is created", loginKey.KeyName))
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreateLoginKey) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	if loginKey, ok := state.GetOk("LoginKey"); ok {
		s.Say("Clean up login key")
		s.Conn.DeleteLoginKey(loginKey.(*LoginKey).KeyName)
	}
}
