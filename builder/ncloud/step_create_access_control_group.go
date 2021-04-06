package ncloud

import (
	"context"
	"fmt"
	"log"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepCreateAccessControlGroup struct {
	Conn                      *NcloudAPIClient
	GetAccessControlGroup     func(acgNo string) (*vserver.AccessControlGroup, error)
	CreateAccessControlGroup  func() (string, error)
	AddAccessControlGroupRule func(acgNo string) error
	Say                       func(message string)
	Error                     func(e error)
	Config                    *Config
	createdAcgNo              string
}

func NewStepCreateAccessControlGroup(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepCreateAccessControlGroup {
	var step = &StepCreateAccessControlGroup{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.GetAccessControlGroup = step.getVpcAccessControlGroup
		step.CreateAccessControlGroup = step.createVpcAccessControlGroup
		step.AddAccessControlGroupRule = step.addVpcAccessControlGroupRule
	}

	return step
}

func (s *StepCreateAccessControlGroup) createVpcAccessControlGroup() (string, error) {
	reqParam := &vserver.CreateAccessControlGroupRequest{
		RegionCode:                    &s.Config.RegionCode,
		AccessControlGroupName:        &s.Config.ServerImageName,
		AccessControlGroupDescription: ncloud.String("Temporary ACG for packer"),
		VpcNo:                         &s.Config.VpcNo,
	}

	resp, err := s.Conn.vserver.V2Api.CreateAccessControlGroup(reqParam)
	if err != nil {
		return "", err
	}

	if resp != nil && *resp.TotalRows > 0 {
		return *resp.AccessControlGroupList[0].AccessControlGroupNo, nil
	}

	return "", nil
}

func (s *StepCreateAccessControlGroup) addVpcAccessControlGroupRule(acgNo string) error {
	_, err := s.Conn.vserver.V2Api.AddAccessControlGroupInboundRule(&vserver.AddAccessControlGroupInboundRuleRequest{
		RegionCode:           &s.Config.RegionCode,
		AccessControlGroupNo: &acgNo,
		VpcNo:                &s.Config.VpcNo,
		AccessControlGroupRuleList: []*vserver.AddAccessControlGroupRuleParameter{
			{
				IpBlock:          ncloud.String("0.0.0.0/0"),
				PortRange:        ncloud.String("22"),
				ProtocolTypeCode: ncloud.String("TCP"),
			},
			{
				IpBlock:          ncloud.String("0.0.0.0/0"),
				PortRange:        ncloud.String("3389"),
				ProtocolTypeCode: ncloud.String("TCP"),
			},
			{
				IpBlock:          ncloud.String("0.0.0.0/0"),
				PortRange:        ncloud.String("5985"),
				ProtocolTypeCode: ncloud.String("TCP"),
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = s.Conn.vserver.V2Api.AddAccessControlGroupOutboundRule(&vserver.AddAccessControlGroupOutboundRuleRequest{
		RegionCode:           &s.Config.RegionCode,
		AccessControlGroupNo: &acgNo,
		VpcNo:                &s.Config.VpcNo,
		AccessControlGroupRuleList: []*vserver.AddAccessControlGroupRuleParameter{
			{
				IpBlock:          ncloud.String("0.0.0.0/0"),
				ProtocolTypeCode: ncloud.String("ICMP"),
			},
			{
				IpBlock:          ncloud.String("0.0.0.0/0"),
				PortRange:        ncloud.String("1-65535"),
				ProtocolTypeCode: ncloud.String("TCP"),
			},
			{
				IpBlock:          ncloud.String("0.0.0.0/0"),
				PortRange:        ncloud.String("1-65535"),
				ProtocolTypeCode: ncloud.String("UDP"),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *StepCreateAccessControlGroup) deleteVpcAccessControlGroup(id string) error {
	reqParam := &vserver.DeleteAccessControlGroupRequest{
		RegionCode:           &s.Config.RegionCode,
		VpcNo:                &s.Config.VpcNo,
		AccessControlGroupNo: ncloud.String(id),
	}

	_, err := s.Conn.vserver.V2Api.DeleteAccessControlGroup(reqParam)
	if err != nil {
		return err
	}

	return nil
}

func (s *StepCreateAccessControlGroup) getVpcAccessControlGroup(id string) (*vserver.AccessControlGroup, error) {
	reqParam := &vserver.GetAccessControlGroupDetailRequest{
		RegionCode:           &s.Config.RegionCode,
		AccessControlGroupNo: ncloud.String(id),
	}

	resp, err := s.Conn.vserver.V2Api.GetAccessControlGroupDetail(reqParam)
	if err != nil {
		return nil, err
	}

	if resp != nil && *resp.TotalRows > 0 {
		return resp.AccessControlGroupList[0], nil
	}

	return nil, nil
}

func (s *StepCreateAccessControlGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create temporary ACG")
	if len(s.Config.AccessControlGroupNo) > 0 {
		acg, err := s.GetAccessControlGroup(s.Config.AccessControlGroupNo)
		if err != nil || acg == nil {
			err := fmt.Errorf("couldn't find specified ACG: %s", err)
			state.Put("error", err)
			return multistep.ActionHalt
		}

		log.Printf("Using specified ACG: %v", s.Config.AccessControlGroupNo)
		return multistep.ActionContinue
	}

	acgNo, err := s.CreateAccessControlGroup()
	s.Say(fmt.Sprintf("Creating temporary ACG [%s]", acgNo))
	if err != nil || len(acgNo) == 0 {
		err := fmt.Errorf("couldn't create ACG for VPC: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	s.createdAcgNo = acgNo

	s.Say(fmt.Sprintf("Creating temporary rules ACG [%s]", acgNo))
	err = s.AddAccessControlGroupRule(acgNo)
	if err != nil {
		err := fmt.Errorf("couldn't create ACG rules for SSH or winrm: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("access_control_group_no", acgNo)

	return processStepResult(err, s.Error, state)
}

func (s *StepCreateAccessControlGroup) Cleanup(state multistep.StateBag) {
	if s.createdAcgNo == "" {
		return
	}

	err := s.deleteVpcAccessControlGroup(s.createdAcgNo)
	if err != nil {
		s.Error(fmt.Errorf("error cleaning up ACG. Please delete the ACG manually: err: %s; ACG No: %s", err, s.createdAcgNo))
	}

	s.Say("Clean up temporary ACG")
}
