package ncloud

import (
	"context"
	"fmt"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
		return nil, err
	}

	return &LoginKey{KeyName, privateKey.PrivateKey}, nil
}

func (s *StepCreateLoginKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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
		s.Conn.DeleteLoginKey(loginKey.(*LoginKey).KeyName)
	}
}
