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

type StepStopServerInstance struct {
	Conn                       *NcloudAPIClient
	StopServerInstance         func(serverInstanceNo string) error
	WaiterServerInstanceStatus func(conn *NcloudAPIClient, serverInstanceNo string, status string, timeout time.Duration) error
	Say                        func(message string)
	Error                      func(e error)
	Config                     *Config
}

func NewStepStopServerInstance(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepStopServerInstance {
	var step = &StepStopServerInstance{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.StopServerInstance = step.stopVpcServerInstance
		step.WaiterServerInstanceStatus = waiterVpcServerInstanceStatus
	} else {
		step.StopServerInstance = step.stopClassicServerInstance
		step.WaiterServerInstanceStatus = waiterClassicServerInstanceStatus
	}

	return step
}

func (s *StepStopServerInstance) stopClassicServerInstance(serverInstanceNo string) error {
	reqParams := &server.StopServerInstancesRequest{
		ServerInstanceNoList: []*string{&serverInstanceNo},
	}

	serverInstanceList, err := s.Conn.server.V2Api.StopServerInstances(reqParams)
	if err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Server Instance is stopping. Server InstanceNo is %s", *serverInstanceList.ServerInstanceList[0].ServerInstanceNo))

	if err := s.WaiterServerInstanceStatus(s.Conn, serverInstanceNo, ServerInstanceStatusStopped, 5*time.Minute); err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Server Instance stopped. Server InstanceNo is %s", *serverInstanceList.ServerInstanceList[0].ServerInstanceNo))

	return nil
}

func (s *StepStopServerInstance) stopVpcServerInstance(serverInstanceNo string) error {
	reqParams := &vserver.StopServerInstancesRequest{
		RegionCode:           &s.Config.RegionCode,
		ServerInstanceNoList: []*string{&serverInstanceNo},
	}

	serverInstanceList, err := s.Conn.vserver.V2Api.StopServerInstances(reqParams)
	if err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Server Instance is stopping. Server InstanceNo is %s", *serverInstanceList.ServerInstanceList[0].ServerInstanceNo))

	if err := s.WaiterServerInstanceStatus(s.Conn, serverInstanceNo, ServerInstanceStatusStopped, 5*time.Minute); err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Server Instance stopped. Server InstanceNo is %s", *serverInstanceList.ServerInstanceList[0].ServerInstanceNo))

	return nil
}

func (s *StepStopServerInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Stop Server Instance")

	var serverInstanceNo = state.Get("instance_no").(string)

	err := s.StopServerInstance(serverInstanceNo)

	return processStepResult(err, s.Error, state)
}

func (*StepStopServerInstance) Cleanup(multistep.StateBag) {
}
