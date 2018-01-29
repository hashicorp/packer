package ncloud

import (
	"context"
	"fmt"
	"log"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreatePublicIPInstance struct {
	Conn                                    *ncloud.Conn
	CreatePublicIPInstance                  func(serverInstanceNo string) (*ncloud.PublicIPInstance, error)
	WaiterAssociatePublicIPToServerInstance func(serverInstanceNo string, publicIP string) error
	Say                                     func(message string)
	Error                                   func(e error)
	Config                                  *Config
}

func NewStepCreatePublicIPInstance(conn *ncloud.Conn, ui packer.Ui, config *Config) *StepCreatePublicIPInstance {
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
	reqParams := new(ncloud.RequestGetServerInstanceList)
	reqParams.ServerInstanceNoList = []string{serverInstanceNo}

	c1 := make(chan error, 1)

	go func() {
		for {
			serverInstanceList, err := s.Conn.GetServerInstanceList(reqParams)

			if err != nil {
				c1 <- err
				return
			}

			if publicIP == serverInstanceList.ServerInstanceList[0].PublicIP {
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

func (s *StepCreatePublicIPInstance) createPublicIPInstance(serverInstanceNo string) (*ncloud.PublicIPInstance, error) {
	reqParams := new(ncloud.RequestCreatePublicIPInstance)
	reqParams.ServerInstanceNo = serverInstanceNo

	publicIPInstanceList, err := s.Conn.CreatePublicIPInstance(reqParams)
	if err != nil {
		return nil, err
	}

	publicIPInstance := publicIPInstanceList.PublicIPInstanceList[0]
	publicIP := publicIPInstance.PublicIP
	s.Say(fmt.Sprintf("Public IP Instance [%s:%s] is created", publicIPInstance.PublicIPInstanceNo, publicIP))

	err = s.waiterAssociatePublicIPToServerInstance(serverInstanceNo, publicIP)

	return &publicIPInstance, nil
}

func (s *StepCreatePublicIPInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create Public IP Instance")

	serverInstanceNo := state.Get("InstanceNo").(string)

	publicIPInstance, err := s.CreatePublicIPInstance(serverInstanceNo)
	if err == nil {
		state.Put("PublicIP", publicIPInstance.PublicIP)
		state.Put("PublicIPInstance", publicIPInstance)
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreatePublicIPInstance) Cleanup(state multistep.StateBag) {
	publicIPInstance, ok := state.GetOk("PublicIPInstance")
	if !ok {
		return
	}

	s.Say("Clean up Public IP Instance")
	publicIPInstanceNo := publicIPInstance.(*ncloud.PublicIPInstance).PublicIPInstanceNo
	s.waitPublicIPInstanceStatus(publicIPInstanceNo, "USED")

	log.Println("Disassociate Public IP Instance ", publicIPInstanceNo)
	s.Conn.DisassociatePublicIP(publicIPInstanceNo)

	s.waitPublicIPInstanceStatus(publicIPInstanceNo, "CREAT")

	reqParams := new(ncloud.RequestDeletePublicIPInstances)
	reqParams.PublicIPInstanceNoList = []string{publicIPInstanceNo}

	log.Println("Delete Public IP Instance ", publicIPInstanceNo)
	s.Conn.DeletePublicIPInstances(reqParams)
}

func (s *StepCreatePublicIPInstance) waitPublicIPInstanceStatus(publicIPInstanceNo string, status string) {
	c1 := make(chan error, 1)

	go func() {
		reqParams := new(ncloud.RequestPublicIPInstanceList)
		reqParams.PublicIPInstanceNoList = []string{publicIPInstanceNo}

		for {
			resp, err := s.Conn.GetPublicIPInstanceList(reqParams)
			if err != nil {
				log.Printf(err.Error())
				c1 <- err
				return
			}

			if resp.TotalRows == 0 {
				c1 <- nil
				return
			}

			instance := resp.PublicIPInstanceList[0]
			if instance.PublicIPInstanceStatus.Code == status && instance.PublicIPInstanceOperation.Code == "NULL" {
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
