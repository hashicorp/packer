package ncloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	PublicIpStatusUsed    = "USED"
	PublicIpStatusCreated = "CREAT"
	PublicIpStatusRunning = "RUN"
)

type StepCreatePublicIP struct {
	Conn                                    *NcloudAPIClient
	GetPublicIP                             func(publicIPNo string) (*server.PublicIpInstance, error)
	CreatePublicIP                          func(serverInstanceNo string) (*server.PublicIpInstance, error)
	DeletePublicIp                          func(publicIPNo string) error
	WaiterAssociatePublicIPToServerInstance func(serverInstanceNo string, publicIP string) error
	DisassociatePublicIpFromServerInstance  func(publicIPNo string) error
	Say                                     func(message string)
	Error                                   func(e error)
	Config                                  *Config
}

func NewStepCreatePublicIP(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepCreatePublicIP {
	var step = &StepCreatePublicIP{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.GetPublicIP = step.getVpcPublicIP
		step.CreatePublicIP = step.createVpcPublicIP
		step.WaiterAssociatePublicIPToServerInstance = step.waiterAssociateVpcPublicIPToServerInstance
		step.DisassociatePublicIpFromServerInstance = step.disassociateVpcPublicIpFromServerInstance
		step.DeletePublicIp = step.deleteVpcPublicIp
	} else {
		step.GetPublicIP = step.getClassicPublicIP
		step.CreatePublicIP = step.createClassicPublicIP
		step.WaiterAssociatePublicIPToServerInstance = step.waiterAssociateClassicPublicIPToServerInstance
		step.DisassociatePublicIpFromServerInstance = step.disassociateClassicPublicIpFromServerInstance
		step.DeletePublicIp = step.deleteClassicPublicIp
	}

	return step
}

func (s *StepCreatePublicIP) getClassicPublicIP(publicIPNo string) (*server.PublicIpInstance, error) {
	reqParams := &server.GetPublicIpInstanceListRequest{
		PublicIpInstanceNoList: []*string{&publicIPNo},
	}

	resp, err := s.Conn.server.V2Api.GetPublicIpInstanceList(reqParams)
	if err != nil {
		return nil, err
	}

	if resp != nil && *resp.TotalRows > 0 {
		return resp.PublicIpInstanceList[0], nil
	}

	return nil, nil
}

func (s *StepCreatePublicIP) getVpcPublicIP(publicIPNo string) (*server.PublicIpInstance, error) {
	reqParams := &vserver.GetPublicIpInstanceDetailRequest{
		RegionCode:         &s.Config.RegionCode,
		PublicIpInstanceNo: &publicIPNo,
	}

	resp, err := s.Conn.vserver.V2Api.GetPublicIpInstanceDetail(reqParams)
	if err != nil {
		return nil, err
	}

	if resp != nil && *resp.TotalRows > 0 {
		inst := resp.PublicIpInstanceList[0]
		return &server.PublicIpInstance{
			PublicIpInstanceNo: inst.PublicIpInstanceNo,
			PublicIp:           inst.PublicIp,
			PublicIpInstanceStatus: &server.CommonCode{
				Code:     inst.PublicIpInstanceStatus.Code,
				CodeName: inst.PublicIpInstanceStatus.CodeName,
			},
			PublicIpInstanceOperation: &server.CommonCode{
				Code:     inst.PublicIpInstanceOperation.Code,
				CodeName: inst.PublicIpInstanceOperation.CodeName,
			},
			ServerInstanceAssociatedWithPublicIp: &server.ServerInstance{
				ServerInstanceNo: inst.ServerInstanceNo,
			},
		}, nil
	}

	return nil, nil
}

func (s *StepCreatePublicIP) waiterAssociateClassicPublicIPToServerInstance(serverInstanceNo string, publicIP string) error {
	reqParams := &server.GetServerInstanceListRequest{
		ServerInstanceNoList: []*string{&serverInstanceNo},
	}

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

func (s *StepCreatePublicIP) waiterAssociateVpcPublicIPToServerInstance(serverInstanceNo string, publicIP string) error {
	reqParams := &vserver.GetServerInstanceDetailRequest{
		RegionCode:       &s.Config.RegionCode,
		ServerInstanceNo: &serverInstanceNo,
	}

	c1 := make(chan error, 1)

	go func() {
		for {
			serverInstanceList, err := s.Conn.vserver.V2Api.GetServerInstanceDetail(reqParams)

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

func (s *StepCreatePublicIP) createClassicPublicIP(serverInstanceNo string) (*server.PublicIpInstance, error) {
	reqParams := &server.CreatePublicIpInstanceRequest{
		ServerInstanceNo: &serverInstanceNo,
	}

	publicIPInstanceList, err := s.Conn.server.V2Api.CreatePublicIpInstance(reqParams)
	if err != nil {
		return nil, err
	}

	publicIPInstance := publicIPInstanceList.PublicIpInstanceList[0]
	publicIP := publicIPInstance.PublicIp
	s.Say(fmt.Sprintf("Public IP Instance [%s:%s] is created", *publicIPInstance.PublicIpInstanceNo, *publicIP))

	err = s.WaiterAssociatePublicIPToServerInstance(serverInstanceNo, *publicIP)
	if err != nil {
		return nil, err
	}

	return publicIPInstance, nil
}

func (s *StepCreatePublicIP) createVpcPublicIP(serverInstanceNo string) (*server.PublicIpInstance, error) {
	reqParams := &vserver.CreatePublicIpInstanceRequest{
		RegionCode:       &s.Config.RegionCode,
		ServerInstanceNo: &serverInstanceNo,
	}

	publicIPInstanceList, err := s.Conn.vserver.V2Api.CreatePublicIpInstance(reqParams)
	if err != nil {
		return nil, err
	}

	publicIPInstance := publicIPInstanceList.PublicIpInstanceList[0]
	publicIP := publicIPInstance.PublicIp
	s.Say(fmt.Sprintf("Public IP Instance [%s:%s] is created", *publicIPInstance.PublicIpInstanceNo, *publicIP))

	err = s.WaiterAssociatePublicIPToServerInstance(serverInstanceNo, *publicIP)
	if err != nil {
		return nil, err
	}

	return &server.PublicIpInstance{
		PublicIpInstanceNo: publicIPInstance.PublicIpInstanceNo,
		PublicIp:           publicIP,
	}, nil
}

func (s *StepCreatePublicIP) waitPublicIPStatus(publicIPInstanceNo string, status string) error {
	c1 := make(chan error, 1)

	go func() {
		for {
			publicIp, err := s.GetPublicIP(publicIPInstanceNo)
			if err != nil {
				c1 <- err
				return
			}

			if publicIp == nil {
				c1 <- nil
				return
			}

			if *publicIp.PublicIpInstanceStatus.Code == status && *publicIp.PublicIpInstanceOperation.Code == "NULL" {
				c1 <- nil
				return
			}

			time.Sleep(time.Second * 2)
		}
	}()

	select {
	case <-c1:
		return nil
	case <-time.After(time.Second * 60):
		return fmt.Errorf("TIMEOUT : wait for public ip[%s] to status[%s] Failed", publicIPInstanceNo, status)
	}
}

func (s *StepCreatePublicIP) disassociateClassicPublicIpFromServerInstance(publicIPInstanceNo string) error {
	reqParams := &server.DisassociatePublicIpFromServerInstanceRequest{
		PublicIpInstanceNo: &publicIPInstanceNo,
	}

	_, err := s.Conn.server.V2Api.DisassociatePublicIpFromServerInstance(reqParams)

	return err
}

func (s *StepCreatePublicIP) disassociateVpcPublicIpFromServerInstance(publicIPInstanceNo string) error {
	reqParams := &vserver.DisassociatePublicIpFromServerInstanceRequest{
		RegionCode:         &s.Config.RegionCode,
		PublicIpInstanceNo: &publicIPInstanceNo,
	}

	_, err := s.Conn.vserver.V2Api.DisassociatePublicIpFromServerInstance(reqParams)

	return err
}

func (s *StepCreatePublicIP) deleteClassicPublicIp(publicIPInstanceNo string) error {
	reqParams := &server.DeletePublicIpInstancesRequest{
		PublicIpInstanceNoList: ncloud.StringList([]string{publicIPInstanceNo}),
	}

	_, err := s.Conn.server.V2Api.DeletePublicIpInstances(reqParams)

	return err
}

func (s *StepCreatePublicIP) deleteVpcPublicIp(publicIPInstanceNo string) error {
	reqParams := &vserver.DeletePublicIpInstanceRequest{
		RegionCode:         &s.Config.RegionCode,
		PublicIpInstanceNo: &publicIPInstanceNo,
	}

	_, err := s.Conn.vserver.V2Api.DeletePublicIpInstance(reqParams)

	return err
}

func (s *StepCreatePublicIP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create Public IP Instance")

	serverInstanceNo := state.Get("instance_no").(string)

	publicIPInstance, err := s.CreatePublicIP(serverInstanceNo)
	if err == nil {
		state.Put("public_ip", *publicIPInstance.PublicIp)
		state.Put("public_ip_instance", publicIPInstance)
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreatePublicIP) Cleanup(state multistep.StateBag) {
	publicIPInstance, ok := state.GetOk("public_ip_instance")
	if !ok {
		return
	}

	s.Say("Clean up Public IP Instance")
	publicIPInstanceNo := publicIPInstance.(*server.PublicIpInstance).PublicIpInstanceNo

	publicIp, err := s.GetPublicIP(*publicIPInstanceNo)
	if err != nil {
		s.Error(err)
	}

	if publicIp == nil {
		return
	}

	publicIpUsedStatus := PublicIpStatusUsed
	publicIpCreatedStatus := PublicIpStatusCreated

	if s.Config.SupportVPC {
		publicIpUsedStatus = PublicIpStatusRunning
		publicIpCreatedStatus = PublicIpStatusRunning
	}

	if publicIp.ServerInstanceAssociatedWithPublicIp != nil &&
		publicIp.ServerInstanceAssociatedWithPublicIp.ServerInstanceNo != nil &&
		len(*publicIp.ServerInstanceAssociatedWithPublicIp.ServerInstanceNo) > 0 {
		if err := s.waitPublicIPStatus(*publicIPInstanceNo, publicIpUsedStatus); err != nil {
			s.Error(err)
		}

		log.Println("Disassociate Public IP Instance ", publicIPInstanceNo)
		if err := s.DisassociatePublicIpFromServerInstance(*publicIPInstanceNo); err != nil {
			s.Error(err)
			return
		}
	}

	if err := s.waitPublicIPStatus(*publicIPInstanceNo, publicIpCreatedStatus); err != nil {
		s.Error(err)
		return
	}

	log.Println("Delete Public IP Instance ", publicIPInstanceNo)
	if err := s.DeletePublicIp(*publicIPInstanceNo); err != nil {
		s.Error(err)
		return
	}
}
