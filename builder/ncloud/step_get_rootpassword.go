package ncloud

import (
	"context"
	"fmt"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepGetRootPassword struct {
	Conn            *ncloud.Conn
	GetRootPassword func(serverInstanceNo string, privateKey string) (string, error)
	Say             func(message string)
	Error           func(e error)
}

func NewStepGetRootPassword(conn *ncloud.Conn, ui packer.Ui) *StepGetRootPassword {
	var step = &StepGetRootPassword{
		Conn:  conn,
		Say:   func(message string) { ui.Say(message) },
		Error: func(e error) { ui.Error(e.Error()) },
	}

	step.GetRootPassword = step.getRootPassword

	return step
}

func (s *StepGetRootPassword) getRootPassword(serverInstanceNo string, privateKey string) (string, error) {
	reqParams := new(ncloud.RequestGetRootPassword)
	reqParams.ServerInstanceNo = serverInstanceNo
	reqParams.PrivateKey = privateKey

	rootPassword, err := s.Conn.GetRootPassword(reqParams)
	if err != nil {
		return "", err
	}

	s.Say(fmt.Sprintf("Root password is %s", rootPassword.RootPassword))

	return rootPassword.RootPassword, nil
}

func (s *StepGetRootPassword) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Get Root Password")

	serverInstanceNo := state.Get("InstanceNo").(string)
	loginKey := state.Get("LoginKey").(*LoginKey)

	rootPassword, err := s.GetRootPassword(serverInstanceNo, loginKey.PrivateKey)

	state.Put("Password", rootPassword)

	return processStepResult(err, s.Error, state)
}

func (*StepGetRootPassword) Cleanup(multistep.StateBag) {
}
