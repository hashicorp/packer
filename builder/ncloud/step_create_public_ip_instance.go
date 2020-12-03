package ncloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepCreatePublicIPInstance struct {
	Conn                                    *NcloudAPIClient
	CreatePublicIPInstance                  func(serverInstanceNo string) (*server.PublicIpInstance, error)
	WaiterAssociatePublicIPToServerInstance func(serverInstanceNo string, publicIP string) error
	Say                                     func(message string)
	Error                                   func(e error)
	Config                                  *Config
}

func NewStepCreatePublicIPInstance(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepCreatePublicIPInstance {
	var step = &StepCreatePublicIPInstance{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	step.CreatePublicIPInstance = step.createPublicIPInstance
	step.WaiterAssociatePublicIPToServerInstance = step.waiterAssociatePublicIPToServerInstance

	return step
}

func (s *StepCreatePublicIPInstance) waiterAssociatePublicIPToServerInstance(serverInstanceNo string, publicIP string) error {
	reqParams := new(server.GetServerInstanceListRequest)
	reqParams.ServerInstanceNoList = []*string{&serverInstanceNo}

	c1 := make(chan error, 1)

	go func() {
		for {
			serverInstanceList, err := s.Conn.server.V2Api.GetServerInstanceList(reqParams)

			if err != nil {
				c1 <- err
				return
			}

			if publicIP == *serverInstanceList.ServerInstanceList[0].PublicIp {
				c1 <- nil
				return
			}

			s.Say("Wait to associate public ip serverInstance")
			time.Sleep(time.Second * 3)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(time.Second * 60):
		return fmt.Errorf("TIMEOUT : association public ip[%s] to server instance[%s] Failed", publicIP, serverInstanceNo)
	}
}

func (s *StepCreatePublicIPInstance) createPublicIPInstance(serverInstanceNo string) (*server.PublicIpInstance, error) {
	reqParams := new(server.CreatePublicIpInstanceRequest)
	reqParams.ServerInstanceNo = &serverInstanceNo

	publicIPInstanceList, err := s.Conn.server.V2Api.CreatePublicIpInstance(reqParams)
	if err != nil {
		return nil, err
	}

	publicIPInstance := publicIPInstanceList.PublicIpInstanceList[0]
	publicIP := publicIPInstance.PublicIp
	s.Say(fmt.Sprintf("Public IP Instance [%s:%s] is created", *publicIPInstance.PublicIpInstanceNo, *publicIP))

	err = s.waiterAssociatePublicIPToServerInstance(serverInstanceNo, *publicIP)
	if err != nil {
		return nil, err
	}

	return publicIPInstance, nil
}

func (s *StepCreatePublicIPInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create Public IP Instance")

	serverInstanceNo := state.Get("InstanceNo").(string)

	publicIPInstance, err := s.CreatePublicIPInstance(serverInstanceNo)
	if err == nil {
		state.Put("PublicIP", *publicIPInstance.PublicIp)
		state.Put("PublicIPInstance", publicIPInstance)
		// instance_id is the generic term used so that users can have access to the
		// instance id inside of the provisioners, used in step_provision.
		state.Put("instance_id", *publicIPInstance)
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreatePublicIPInstance) Cleanup(state multistep.StateBag) {
	publicIPInstance, ok := state.GetOk("PublicIPInstance")
	if !ok {
		return
	}

	s.Say("Clean up Public IP Instance")
	publicIPInstanceNo := publicIPInstance.(*server.PublicIpInstance).PublicIpInstanceNo
	s.waitPublicIPInstanceStatus(publicIPInstanceNo, "USED")

	log.Println("Disassociate Public IP Instance ", publicIPInstanceNo)
	reqParams := &server.DisassociatePublicIpFromServerInstanceRequest{PublicIpInstanceNo: publicIPInstanceNo}
	s.Conn.server.V2Api.DisassociatePublicIpFromServerInstance(reqParams)

	s.waitPublicIPInstanceStatus(publicIPInstanceNo, "CREAT")

	reqDeleteParams := &server.DeletePublicIpInstancesRequest{
		PublicIpInstanceNoList: ncloud.StringList([]string{*publicIPInstanceNo}),
	}

	log.Println("Delete Public IP Instance ", publicIPInstanceNo)
	s.Conn.server.V2Api.DeletePublicIpInstances(reqDeleteParams)
}

func (s *StepCreatePublicIPInstance) waitPublicIPInstanceStatus(publicIPInstanceNo *string, status string) {
	c1 := make(chan error, 1)

	go func() {
		reqParams := new(server.GetPublicIpInstanceListRequest)
		reqParams.PublicIpInstanceNoList = []*string{publicIPInstanceNo}

		for {
			resp, err := s.Conn.server.V2Api.GetPublicIpInstanceList(reqParams)
			if err != nil {
				log.Printf(err.Error())
				c1 <- err
				return
			}

			if *resp.TotalRows == 0 {
				c1 <- nil
				return
			}

			instance := resp.PublicIpInstanceList[0]
			if *instance.PublicIpInstanceStatus.Code == status && *instance.PublicIpInstanceOperation.Code == "NULL" {
				c1 <- nil
				return
			}

			time.Sleep(time.Second * 2)
		}
	}()

	select {
	case <-c1:
		return
	case <-time.After(time.Second * 60):
		return
	}
}
