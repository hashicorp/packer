package jdcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
)

type stepValidateParameters struct {
	InstanceSpecConfig *JDCloudInstanceSpecConfig
	ui                 packer.Ui
	state              multistep.StateBag
}

func (s *stepValidateParameters) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	s.ui = state.Get("ui").(packer.Ui)
	s.state = state
	s.ui.Say("Validating parameters...")

	if err := s.ValidateSubnetFunc(); err != nil {
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := s.ValidateImageFunc(); err != nil {
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepValidateParameters) ValidateSubnetFunc() error {

	subnetId := s.InstanceSpecConfig.SubnetId
	if len(subnetId) == 0 {
		s.ui.Message("\t 'subnet' is not specified, we will create a new one for you :) ")
		return s.CreateRandomSubnet()
	}

	s.ui.Message("\t validating your subnet:" + s.InstanceSpecConfig.SubnetId)
	req := vpc.NewDescribeSubnetRequest(Region, subnetId)
	resp, err := VpcClient.DescribeSubnet(req)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed in validating subnet->%s, reasons:%v", subnetId, err)
	}
	if resp != nil && resp.Error.Code != FINE {
		return fmt.Errorf("[ERROR] Something wrong with your subnet->%s, reasons:%v", subnetId, resp.Error)
	}

	s.ui.Message("\t subnet found:" + resp.Result.Subnet.SubnetName)
	return nil

}

func (s *stepValidateParameters) ValidateImageFunc() error {

	s.ui.Message("\t validating your base image:" + s.InstanceSpecConfig.ImageId)
	imageId := s.InstanceSpecConfig.ImageId
	req := vm.NewDescribeImageRequest(Region, imageId)
	resp, err := VmClient.DescribeImage(req)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed in validating your image->%s, reasons:%v", imageId, err)
	}
	if resp != nil && resp.Error.Code != FINE {
		return fmt.Errorf("[ERROR] Something wrong with your image->%s, reasons:%v", imageId, resp.Error)
	}

	s.ui.Message("\t image found:" + resp.Result.Image.Name)
	s.state.Put("source_image", &resp.Result.Image)
	return nil
}

func (s *stepValidateParameters) CreateRandomSubnet() error {

	newVpc, err := s.CreateRandomVpc()
	if err != nil {
		return err
	}

	req := vpc.NewCreateSubnetRequest(Region, newVpc, "created_by_packer", "192.168.0.0/20")
	resp, err := VpcClient.CreateSubnet(req)
	if err != nil || resp.Error.Code != FINE {
		errorMessage := fmt.Sprintf("[ERROR] Failed in creating new subnet :( \n error:%v \n response:%v", err, resp)
		s.ui.Error(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	s.InstanceSpecConfig.SubnetId = resp.Result.SubnetId
	s.ui.Message("\t\t Hi, we have created a new subnet for you :) its name is 'created_by_packer' and its id=" + resp.Result.SubnetId)
	return nil
}

func (s *stepValidateParameters) CreateRandomVpc() (string, error) {
	req := vpc.NewCreateVpcRequest(Region, "created_by_packer")
	resp, err := VpcClient.CreateVpc(req)
	if err != nil || resp.Error.Code != FINE {
		errorMessage := fmt.Sprintf("[ERROR] Failed in creating new vpc :( \n error :%v, \n response:%v", err, resp)
		s.ui.Error(errorMessage)
		return "", fmt.Errorf(errorMessage)
	}
	return resp.Result.VpcId, nil
}

func (s *stepValidateParameters) Cleanup(state multistep.StateBag) {}
