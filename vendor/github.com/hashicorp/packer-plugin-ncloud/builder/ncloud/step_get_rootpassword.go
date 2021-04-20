package ncloud

import (
	"context"
	"fmt"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepGetRootPassword struct {
	Conn            *NcloudAPIClient
	GetRootPassword func(serverInstanceNo string, privateKey string) (string, error)
	Say             func(message string)
	Error           func(e error)
	Config          *Config
}

func NewStepGetRootPassword(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepGetRootPassword {
	var step = &StepGetRootPassword{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.GetRootPassword = step.getVpcRootPassword
	} else {
		step.GetRootPassword = step.getClassicRootPassword
	}

	return step
}

func (s *StepGetRootPassword) getClassicRootPassword(serverInstanceNo string, privateKey string) (string, error) {
	reqParams := &server.GetRootPasswordRequest{
		ServerInstanceNo: &serverInstanceNo,
		PrivateKey:       &privateKey,
	}

	rootPassword, err := s.Conn.server.V2Api.GetRootPassword(reqParams)
	if err != nil {
		return "", err
	}

	s.Say(fmt.Sprintf("Root password is %s", *rootPassword.RootPassword))

	return *rootPassword.RootPassword, nil
}

func (s *StepGetRootPassword) getVpcRootPassword(serverInstanceNo string, privateKey string) (string, error) {
	reqParams := &vserver.GetRootPasswordRequest{
		RegionCode:       &s.Config.RegionCode,
		ServerInstanceNo: &serverInstanceNo,
		PrivateKey:       &privateKey,
	}

	rootPassword, err := s.Conn.vserver.V2Api.GetRootPassword(reqParams)
	if err != nil {
		return "", err
	}

	s.Say(fmt.Sprintf("Root password is %s", *rootPassword.RootPassword))

	return *rootPassword.RootPassword, nil
}

func (s *StepGetRootPassword) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Get Root Password")

	serverInstanceNo := state.Get("instance_no").(string)
	loginKey := state.Get("login_key").(*LoginKey)

	rootPassword, err := s.GetRootPassword(serverInstanceNo, loginKey.PrivateKey)

	if s.Config.Comm.Type == "ssh" {
		s.Config.Comm.SSHPassword = rootPassword
	} else if s.Config.Comm.Type == "winrm" {
		s.Config.Comm.WinRMPassword = rootPassword
	}

	return processStepResult(err, s.Error, state)
}

func (*StepGetRootPassword) Cleanup(multistep.StateBag) {
}
