package ncloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepTerminateServerInstance struct {
	Conn                    *NcloudAPIClient
	TerminateServerInstance func(serverInstanceNo string) error
	Say                     func(message string)
	Error                   func(e error)
	Config                  *Config
}

func NewStepTerminateServerInstance(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepTerminateServerInstance {
	var step = &StepTerminateServerInstance{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.TerminateServerInstance = step.terminateVpcServerInstance
	} else {
		step.TerminateServerInstance = step.terminateClassicServerInstance
	}

	return step
}

func (s *StepTerminateServerInstance) terminateClassicServerInstance(serverInstanceNo string) error {
	reqParams := &server.TerminateServerInstancesRequest{
		ServerInstanceNoList: []*string{&serverInstanceNo},
	}

	_, err := s.Conn.server.V2Api.TerminateServerInstances(reqParams)
	if err != nil {
		return err
	}
	s.Say(fmt.Sprintf("Server Instance is terminating. Server InstanceNo is %s", serverInstanceNo))

	c1 := make(chan error, 1)

	go func() {
		reqParams := &server.GetServerInstanceListRequest{
			ServerInstanceNoList: []*string{&serverInstanceNo},
		}

		for {
			serverInstanceList, err := s.Conn.server.V2Api.GetServerInstanceList(reqParams)
			if err != nil {
				c1 <- err
				return
			} else if *serverInstanceList.TotalRows == 0 {
				c1 <- nil
				return
			}

			log.Printf("Wating for terminating server instance [%s] is %s\n", serverInstanceNo, *serverInstanceList.ServerInstanceList[0].ServerInstanceStatus.Code)
			time.Sleep(time.Second * 3)
		}
	}()

	select {
	case res := <-c1:
		s.Say(fmt.Sprintf("Server Instance terminated. Server InstanceNo is %s", serverInstanceNo))
		return res
	case <-time.After(time.Second * 60):
		return errors.New("TIMEOUT : Can't terminate server instance")
	}
}

func (s *StepTerminateServerInstance) terminateVpcServerInstance(serverInstanceNo string) error {
	reqParams := &vserver.TerminateServerInstancesRequest{
		ServerInstanceNoList: []*string{&serverInstanceNo},
	}

	_, err := s.Conn.vserver.V2Api.TerminateServerInstances(reqParams)
	if err != nil {
		return err
	}
	s.Say(fmt.Sprintf("Server Instance is terminating. Server InstanceNo is %s", serverInstanceNo))

	c1 := make(chan error, 1)

	go func() {
		reqParams := &vserver.GetServerInstanceListRequest{
			RegionCode:           &s.Config.RegionCode,
			ServerInstanceNoList: []*string{&serverInstanceNo},
		}

		for {
			serverInstanceList, err := s.Conn.vserver.V2Api.GetServerInstanceList(reqParams)
			if err != nil {
				c1 <- err
				return
			} else if *serverInstanceList.TotalRows == 0 {
				c1 <- nil
				return
			}

			log.Printf("Wating for terminating server instance [%s] is %s\n", serverInstanceNo, *serverInstanceList.ServerInstanceList[0].ServerInstanceStatus.Code)
			time.Sleep(time.Second * 3)
		}
	}()

	select {
	case res := <-c1:
		s.Say(fmt.Sprintf("Server Instance terminated. Server InstanceNo is %s", serverInstanceNo))
		return res
	case <-time.After(time.Second * 120):
		return errors.New("TIMEOUT : Can't terminate server instance")
	}
}

func (s *StepTerminateServerInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Terminate Server Instance")

	var serverInstanceNo = state.Get("instance_no").(string)

	err := s.TerminateServerInstance(serverInstanceNo)

	return processStepResult(err, s.Error, state)
}

func (*StepTerminateServerInstance) Cleanup(multistep.StateBag) {
}
