package ncloud

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	ServerInstanceStatusStopped    = "NSTOP"
	ServerInstanceStatusTerminated = "TERMT"
	ServerInstanceStatusRunning    = "RUN"
)

type StepCreateServerInstance struct {
	Conn                       *NcloudAPIClient
	CreateServerInstance       func(loginKeyName string, feeSystemTypeCode string, state multistep.StateBag) (string, error)
	GetServerInstance          func() (string, string, error)
	WaiterServerInstanceStatus func(conn *NcloudAPIClient, serverInstanceNo string, status string, timeout time.Duration) error
	Say                        func(message string)
	Error                      func(e error)
	Config                     *Config
	serverInstanceNo           string
}

func NewStepCreateServerInstance(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepCreateServerInstance {
	var step = &StepCreateServerInstance{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.CreateServerInstance = step.createVpcServerInstance
		step.WaiterServerInstanceStatus = waiterVpcServerInstanceStatus
		step.GetServerInstance = step.getVpcServerInstance
	} else {
		step.CreateServerInstance = step.createClassicServerInstance
		step.WaiterServerInstanceStatus = waiterClassicServerInstanceStatus
		step.GetServerInstance = step.getClassicServerInstance
	}

	return step
}

func (s *StepCreateServerInstance) createClassicServerInstance(loginKeyName string, feeSystemTypeCode string, state multistep.StateBag) (string, error) {
	var zoneNo = state.Get("zone_no").(string)

	reqParams := &server.CreateServerInstancesRequest{
		ServerProductCode:   &s.Config.ServerProductCode,
		MemberServerImageNo: &s.Config.MemberServerImageNo,
		LoginKeyName:        &loginKeyName,
		ZoneNo:              &zoneNo,
		FeeSystemTypeCode:   &feeSystemTypeCode,
	}
	if s.Config.MemberServerImageNo == "" {
		reqParams.ServerImageProductCode = &s.Config.ServerImageProductCode
	}

	if s.Config.UserData != "" {
		reqParams.UserData = &s.Config.UserData
	}

	if s.Config.UserDataFile != "" {
		contents, err := ioutil.ReadFile(s.Config.UserDataFile)
		if err != nil {
			return "", fmt.Errorf("Problem reading user data file: %s", err)
		}

		reqParams.UserData = ncloud.String(string(contents))
	}

	if s.Config.AccessControlGroupNo != "" {
		reqParams.AccessControlGroupConfigurationNoList = []*string{&s.Config.AccessControlGroupNo}
	}

	serverInstanceList, err := s.Conn.server.V2Api.CreateServerInstances(reqParams)
	if err != nil {
		return "", err
	}

	s.serverInstanceNo = *serverInstanceList.ServerInstanceList[0].ServerInstanceNo
	s.Say(fmt.Sprintf("Server Instance is creating. Server InstanceNo is %s", s.serverInstanceNo))
	log.Println("Server Instance information : ", serverInstanceList.ServerInstanceList[0])

	if err := s.WaiterServerInstanceStatus(s.Conn, s.serverInstanceNo, ServerInstanceStatusRunning, 30*time.Minute); err != nil {
		return "", errors.New("TIMEOUT : server instance status is not running")
	}

	s.Say(fmt.Sprintf("Server Instance is created. Server InstanceNo is %s", s.serverInstanceNo))

	return s.serverInstanceNo, nil
}

func (s *StepCreateServerInstance) createVpcServerInstance(loginKeyName string, feeSystemTypeCode string, state multistep.StateBag) (string, error) {
	var initScriptNo string
	var acgNo string
	var err error

	if v, ok := state.GetOk("init_script_no"); ok {
		initScriptNo = v.(string)
	}

	if s.Config.AccessControlGroupNo != "" {
		acgNo = s.Config.AccessControlGroupNo
	} else {
		acgNo = state.Get("access_control_group_no").(string)
	}

	reqParams := &vserver.CreateServerInstancesRequest{
		RegionCode:                  &s.Config.RegionCode,
		ServerProductCode:           &s.Config.ServerProductCode,
		MemberServerImageInstanceNo: &s.Config.MemberServerImageNo,
		LoginKeyName:                &loginKeyName,
		FeeSystemTypeCode:           &feeSystemTypeCode,
		InitScriptNo:                &initScriptNo,
		VpcNo:                       &s.Config.VpcNo,
		SubnetNo:                    &s.Config.SubnetNo,
		NetworkInterfaceList: []*vserver.NetworkInterfaceParameter{{
			NetworkInterfaceOrder:    ncloud.Int32(0),
			AccessControlGroupNoList: []*string{ncloud.String(acgNo)}},
		},
	}

	if s.Config.MemberServerImageNo == "" {
		reqParams.ServerImageProductCode = &s.Config.ServerImageProductCode
	}

	serverInstanceList, err := s.Conn.vserver.V2Api.CreateServerInstances(reqParams)
	if err != nil {
		return "", err
	}

	s.serverInstanceNo = *serverInstanceList.ServerInstanceList[0].ServerInstanceNo
	s.Say(fmt.Sprintf("Server Instance is creating. Server InstanceNo is %s", s.serverInstanceNo))
	log.Println("Server Instance information : ", serverInstanceList.ServerInstanceList[0])

	if err := s.WaiterServerInstanceStatus(s.Conn, s.serverInstanceNo, ServerInstanceStatusRunning, 30*time.Minute); err != nil {
		return "", errors.New("TIMEOUT : server instance status is not running")
	}

	s.Say(fmt.Sprintf("Server Instance is created. Server InstanceNo is %s", s.serverInstanceNo))

	return s.serverInstanceNo, nil
}

func (s *StepCreateServerInstance) getClassicServerInstance() (string, string, error) {
	reqParams := &server.GetServerInstanceListRequest{
		ServerInstanceNoList: []*string{&s.serverInstanceNo},
	}

	resp, err := s.Conn.server.V2Api.GetServerInstanceList(reqParams)
	if err != nil {
		return "", "", err
	}

	if *resp.TotalRows > 0 {
		return *resp.ServerInstanceList[0].ServerInstanceNo, *resp.ServerInstanceList[0].ServerInstanceStatus.Code, nil
	} else {
		return "", "", nil
	}
}

func (s *StepCreateServerInstance) getVpcServerInstance() (string, string, error) {
	reqParams := &vserver.GetServerInstanceDetailRequest{
		ServerInstanceNo: &s.serverInstanceNo,
	}

	resp, err := s.Conn.vserver.V2Api.GetServerInstanceDetail(reqParams)
	if err != nil {
		return "", "", err
	}

	if *resp.TotalRows > 0 {
		return *resp.ServerInstanceList[0].ServerInstanceNo, *resp.ServerInstanceList[0].ServerInstanceStatus.Code, nil
	} else {
		return "", "", nil
	}
}

func (s *StepCreateServerInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create Server Instance")
	loginKey := state.Get("login_key").(*LoginKey)

	feeSystemTypeCode := "MTRAT"
	if _, ok := state.GetOk("fee_system_type_code"); ok {
		feeSystemTypeCode = state.Get("fee_system_type_code").(string)
	}

	serverInstanceNo, err := s.CreateServerInstance(loginKey.KeyName, feeSystemTypeCode, state)
	if err == nil {
		state.Put("instance_no", serverInstanceNo)
		// instance_id is the generic term used so that users can have access to the
		// instance id inside of the provisioners, used in step_provision.
		state.Put("instance_id", serverInstanceNo)
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreateServerInstance) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	ui := state.Get("ui").(packersdk.Ui)

	if !cancelled && !halted {
		return
	}

	if s.serverInstanceNo == "" {
		return
	}

	_, status, err := s.GetServerInstance()

	if err != nil || len(status) == 0 {
		return
	}

	s.Say("Clean up Server Instance")

	// stop server instance
	if status != ServerInstanceStatusStopped && status != ServerInstanceStatusTerminated {
		stepStopServerInstance := NewStepStopServerInstance(s.Conn, ui, s.Config)
		err := stepStopServerInstance.StopServerInstance(s.serverInstanceNo)
		if err != nil {
			s.Error(err)
			return
		}
	}

	// terminate server instance
	if status != ServerInstanceStatusTerminated {
		stepStopServerInstance := NewStepTerminateServerInstance(s.Conn, ui, s.Config)
		err := stepStopServerInstance.TerminateServerInstance(s.serverInstanceNo)
		if err != nil {
			s.Error(err)
			return
		}
	}
}
