package cvm

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepRunInstance struct {
	InstanceType             string
	UserData                 string
	UserDataFile             string
	instanceId               string
	ZoneId                   string
	InstanceName             string
	DiskType                 string
	DiskSize                 int64
	HostName                 string
	InternetMaxBandwidthOut  int64
	AssociatePublicIpAddress bool
}

func (s *stepRunInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("cvm_client").(*cvm.Client)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	source_image := state.Get("source_image").(*cvm.Image)
	vpc_id := state.Get("vpc_id").(string)
	subnet_id := state.Get("subnet_id").(string)
	security_group_id := state.Get("security_group_id").(string)

	password := config.Comm.SSHPassword
	if password == "" && config.Comm.WinRMPassword != "" {
		password = config.Comm.WinRMPassword
	}
	userData, err := s.getUserData(state)
	if err != nil {
		err := fmt.Errorf("get user_data failed: %s", err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Creating Instance.")
	// config RunInstances parameters
	POSTPAID_BY_HOUR := "POSTPAID_BY_HOUR"
	req := cvm.NewRunInstancesRequest()
	if s.ZoneId != "" {
		req.Placement = &cvm.Placement{
			Zone: &s.ZoneId,
		}
	}
	req.ImageId = source_image.ImageId
	req.InstanceChargeType = &POSTPAID_BY_HOUR
	req.InstanceType = &s.InstanceType
	req.SystemDisk = &cvm.SystemDisk{
		DiskType: &s.DiskType,
		DiskSize: &s.DiskSize,
	}
	req.VirtualPrivateCloud = &cvm.VirtualPrivateCloud{
		VpcId:    &vpc_id,
		SubnetId: &subnet_id,
	}
	TRAFFIC_POSTPAID_BY_HOUR := "TRAFFIC_POSTPAID_BY_HOUR"
	if s.AssociatePublicIpAddress {
		req.InternetAccessible = &cvm.InternetAccessible{
			InternetChargeType:      &TRAFFIC_POSTPAID_BY_HOUR,
			InternetMaxBandwidthOut: &s.InternetMaxBandwidthOut,
		}
	}
	req.InstanceName = &s.InstanceName
	loginSettings := cvm.LoginSettings{}
	if password != "" {
		loginSettings.Password = &password
	}
	if config.Comm.SSHKeyPairName != "" {
		loginSettings.KeyIds = []*string{&config.Comm.SSHKeyPairName}
	}
	req.LoginSettings = &loginSettings
	req.SecurityGroupIds = []*string{&security_group_id}
	req.ClientToken = &s.InstanceName
	req.HostName = &s.HostName
	req.UserData = &userData

	resp, err := client.RunInstances(req)
	if err != nil {
		err := fmt.Errorf("create instance failed: %s", err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if len(resp.Response.InstanceIdSet) != 1 {
		err := fmt.Errorf("create instance failed: %d instance(s) created", len(resp.Response.InstanceIdSet))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.instanceId = *resp.Response.InstanceIdSet[0]

	err = WaitForInstance(client, s.instanceId, "RUNNING", 1800)
	if err != nil {
		err := fmt.Errorf("wait instance launch failed: %s", err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	describeReq := cvm.NewDescribeInstancesRequest()
	describeReq.InstanceIds = []*string{&s.instanceId}
	describeResp, err := client.DescribeInstances(describeReq)
	if err != nil {
		err := fmt.Errorf("wait instance launch failed: %s", err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("instance", describeResp.Response.InstanceSet[0])
	return multistep.ActionContinue
}

func (s *stepRunInstance) getUserData(state multistep.StateBag) (string, error) {
	userData := s.UserData
	if userData == "" && s.UserDataFile != "" {
		data, err := ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			return "", err
		}
		userData = string(data)
	}
	userData = base64.StdEncoding.EncodeToString([]byte(userData))
	log.Printf(fmt.Sprintf("user_data: %s", userData))
	return userData, nil
}

func (s *stepRunInstance) Cleanup(state multistep.StateBag) {
	if s.instanceId == "" {
		return
	}
	MessageClean(state, "instance")
	client := state.Get("cvm_client").(*cvm.Client)
	ui := state.Get("ui").(packer.Ui)
	req := cvm.NewTerminateInstancesRequest()
	req.InstanceIds = []*string{&s.instanceId}
	_, err := client.TerminateInstances(req)
	// The binding relation between instance and vpc would last few minutes after
	// instance terminate, we sleep here to give more time
	time.Sleep(2 * time.Minute)
	if err != nil {
		ui.Error(fmt.Sprintf("terminate instance(%s) failed: %s", s.instanceId, err.Error()))
	}
}
