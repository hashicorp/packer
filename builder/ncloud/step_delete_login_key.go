package ncloud

import (
	"fmt"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type StepDeleteLoginKey struct {
	Conn           *ncloud.Conn
	DeleteLoginKey func(keyName string) error
	Say            func(message string)
	Error          func(e error)
}

func NewStepDeleteLoginKey(conn *ncloud.Conn, ui packer.Ui) *StepDeleteLoginKey {
	var step = &StepDeleteLoginKey{
		Conn:  conn,
		Say:   func(message string) { ui.Say(message) },
		Error: func(e error) { ui.Error(e.Error()) },
	}

	step.DeleteLoginKey = step.deleteLoginKey

	return step
}

func (s *StepDeleteLoginKey) deleteLoginKey(keyName string) error {
	resp, err := s.Conn.DeleteLoginKey(keyName)
	if err != nil {
		return err
	}

	return nil
}

func (s *StepDeleteLoginKey) Run(state multistep.StateBag) multistep.StepAction {
	var loginKey = state.Get("LoginKey").(*LoginKey)

	err := s.DeleteLoginKey(loginKey.KeyName)
	if err == nil {
		s.Say(fmt.Sprintf("Login Key[%s] is deleted", loginKey.KeyName))
	}

	return processStepResult(err, s.Error, state)
}

func (*StepDeleteLoginKey) Cleanup(multistep.StateBag) {
}
