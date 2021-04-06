package ncloud

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type InitScript struct {
	KeyName    string
	PrivateKey string
}

type StepCreateInitScript struct {
	Conn             *NcloudAPIClient
	CreateInitScript func() (string, error)
	DeleteInitScript func(initScriptNo string) error
	Say              func(message string)
	Error            func(e error)
	Config           *Config
}

func NewStepCreateInitScript(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepCreateInitScript {
	var step = &StepCreateInitScript{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.CreateInitScript = step.createVpcInitScript
		step.DeleteInitScript = step.deleteVpcInitScript
	}

	return step
}

func (s *StepCreateInitScript) createVpcInitScript() (string, error) {
	name := fmt.Sprintf("packer-%d", time.Now().Unix())
	reqParams := &vserver.CreateInitScriptRequest{
		RegionCode:     &s.Config.RegionCode,
		InitScriptName: &name,
	}

	if s.Config.Comm.Type == "winrm" {
		reqParams.OsTypeCode = ncloud.String("WND")
	}

	if s.Config.UserData != "" {
		reqParams.InitScriptContent = &s.Config.UserData
	}
	if s.Config.UserDataFile != "" {
		contents, err := ioutil.ReadFile(s.Config.UserDataFile)
		if err != nil {
			return "", fmt.Errorf("Problem reading user data file: %s", err)
		}

		reqParams.InitScriptContent = ncloud.String(string(contents))
	}

	if reqParams.InitScriptContent == nil {
		return "", nil
	}

	resp, err := s.Conn.vserver.V2Api.CreateInitScript(reqParams)
	if err != nil {
		return "", err
	}

	return *resp.InitScriptList[0].InitScriptNo, nil
}

func (s *StepCreateInitScript) deleteVpcInitScript(initScriptNo string) error {
	reqParams := &vserver.DeleteInitScriptsRequest{
		RegionCode:       &s.Config.RegionCode,
		InitScriptNoList: []*string{&initScriptNo},
	}
	_, err := s.Conn.vserver.V2Api.DeleteInitScripts(reqParams)

	if err != nil {
		return err
	}

	return nil
}

func (s *StepCreateInitScript) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.Config.UserData) == 0 && len(s.Config.UserDataFile) == 0 {
		return multistep.ActionContinue
	}
	s.Say("Create Init script")
	initScriptNo, err := s.CreateInitScript()
	if err == nil && initScriptNo != "" {
		state.Put("init_script_no", initScriptNo)
		s.Say(fmt.Sprintf("Init script[%s] is created", initScriptNo))
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreateInitScript) Cleanup(state multistep.StateBag) {
	if initScriptNo, ok := state.GetOk("init_script_no"); ok {
		s.Say("Cleanup Init script")
		if err := s.DeleteInitScript(initScriptNo.(string)); err != nil {
			s.Error(err)
			return
		}
	}
}
