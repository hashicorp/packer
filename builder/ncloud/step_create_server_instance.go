package ncloud

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateServerInstance struct {
	Conn                               *ncloud.Conn
	CreateServerInstance               func(loginKeyName string, zoneNo string, feeSystemTypeCode string) (string, error)
	CheckServerInstanceStatusIsRunning func(serverInstanceNo string) error
	Say                                func(message string)
	Error                              func(e error)
	Config                             *Config
	serverInstanceNo                   string
}

func NewStepCreateServerInstance(conn *ncloud.Conn, ui packer.Ui, config *Config) *StepCreateServerInstance {
	var step = &StepCreateServerInstance{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	step.CreateServerInstance = step.createServerInstance

	return step
}

func (s *StepCreateServerInstance) createServerInstance(loginKeyName string, zoneNo string, feeSystemTypeCode string) (string, error) {
	reqParams := new(ncloud.RequestCreateServerInstance)
	reqParams.ServerProductCode = s.Config.ServerProductCode
	reqParams.MemberServerImageNo = s.Config.MemberServerImageNo
	if s.Config.MemberServerImageNo == "" {
		reqParams.ServerImageProductCode = s.Config.ServerImageProductCode
	}
	reqParams.LoginKeyName = loginKeyName
	reqParams.ZoneNo = zoneNo
	reqParams.FeeSystemTypeCode = feeSystemTypeCode

	if s.Config.UserData != "" {
		reqParams.UserData = s.Config.UserData
	}

	if s.Config.UserDataFile != "" {
		contents, err := ioutil.ReadFile(s.Config.UserDataFile)
		if err != nil {
			return "", fmt.Errorf("Problem reading user data file: %s", err)
		}

		reqParams.UserData = string(contents)
	}

	if s.Config.AccessControlGroupConfigurationNo != "" {
		reqParams.AccessControlGroupConfigurationNoList = []string{s.Config.AccessControlGroupConfigurationNo}
	}

	serverInstanceList, err := s.Conn.CreateServerInstances(reqParams)
	if err != nil {
		return "", err
	}

	s.serverInstanceNo = serverInstanceList.ServerInstanceList[0].ServerInstanceNo
	s.Say(fmt.Sprintf("Server Instance is creating. Server InstanceNo is %s", s.serverInstanceNo))
	log.Println("Server Instance information : ", serverInstanceList.ServerInstanceList[0])

	if err := waiterServerInstanceStatus(s.Conn, s.serverInstanceNo, "RUN", 30*time.Minute); err != nil {
		return "", errors.New("TIMEOUT : server instance status is not running")
	}

	s.Say(fmt.Sprintf("Server Instance is created. Server InstanceNo is %s", s.serverInstanceNo))

	return s.serverInstanceNo, nil
}

func (s *StepCreateServerInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create Server Instance")

	var loginKey = state.Get("LoginKey").(*LoginKey)
	var zoneNo = state.Get("ZoneNo").(string)

	feeSystemTypeCode := "MTRAT"
	if _, ok := state.GetOk("FeeSystemTypeCode"); ok {
		feeSystemTypeCode = state.Get("FeeSystemTypeCode").(string)
	}

	serverInstanceNo, err := s.CreateServerInstance(loginKey.KeyName, zoneNo, feeSystemTypeCode)
	if err == nil {
		state.Put("InstanceNo", serverInstanceNo)
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreateServerInstance) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	if s.serverInstanceNo == "" {
		return
	}

	reqParams := new(ncloud.RequestGetServerInstanceList)
	reqParams.ServerInstanceNoList = []string{s.serverInstanceNo}

	serverInstanceList, err := s.Conn.GetServerInstanceList(reqParams)
	if err != nil || serverInstanceList.TotalRows == 0 {
		return
	}

	s.Say("Clean up Server Instance")

	serverInstance := serverInstanceList.ServerInstanceList[0]
	// stop server instance
	if serverInstance.ServerInstanceStatus.Code != "NSTOP" && serverInstance.ServerInstanceStatus.Code != "TERMT" {
		reqParams := new(ncloud.RequestStopServerInstances)
		reqParams.ServerInstanceNoList = []string{s.serverInstanceNo}

		log.Println("Stop Server Instance")
		s.Conn.StopServerInstances(reqParams)
		waiterServerInstanceStatus(s.Conn, s.serverInstanceNo, "NSTOP", time.Minute)
	}

	// terminate server instance
	if serverInstance.ServerInstanceStatus.Code != "TERMT" {
		reqParams := new(ncloud.RequestTerminateServerInstances)
		reqParams.ServerInstanceNoList = []string{s.serverInstanceNo}

		log.Println("Terminate Server Instance")
		s.Conn.TerminateServerInstances(reqParams)
	}
}
