package jdcloud

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	vpcApis "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	vpcClient "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
)

type stepCreateJDCloudInstance struct {
	InstanceSpecConfig *JDCloudInstanceSpecConfig
	CredentialConfig   *JDCloudCredentialConfig
	ui                 packer.Ui
}

func (s *stepCreateJDCloudInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	privateKey := s.InstanceSpecConfig.Comm.SSHPrivateKey
	keyName := s.InstanceSpecConfig.Comm.SSHKeyPairName
	password := s.InstanceSpecConfig.Comm.SSHPassword
	s.ui = state.Get("ui").(packer.Ui)
	s.ui.Say("Creating instances")

	instanceSpec := vm.InstanceSpec{
		Az:           &s.CredentialConfig.Az,
		InstanceType: &s.InstanceSpecConfig.InstanceType,
		ImageId:      &s.InstanceSpecConfig.ImageId,
		Name:         s.InstanceSpecConfig.InstanceName,
		PrimaryNetworkInterface: &vm.InstanceNetworkInterfaceAttachmentSpec{
			NetworkInterface: &vpc.NetworkInterfaceSpec{
				SubnetId: s.InstanceSpecConfig.SubnetId,
				Az:       &s.CredentialConfig.Az,
			},
		},
	}

	if len(password) > 0 {
		instanceSpec.Password = &password
	}
	if len(keyName) > 0 && len(privateKey) > 0 {
		instanceSpec.KeyNames = []string{keyName}
	}

	req := apis.NewCreateInstancesRequest(Region, &instanceSpec)
	resp, err := VmClient.CreateInstances(req)

	if err != nil || resp.Error.Code != FINE {
		err := fmt.Errorf("Error creating instance, error-%v response:%v", err, resp)
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.InstanceSpecConfig.InstanceId = resp.Result.InstanceIds[0]
	instanceInterface, err := InstanceStatusRefresher(s.InstanceSpecConfig.InstanceId, []string{VM_PENDING, VM_STARTING}, []string{VM_RUNNING})
	if err != nil {
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}

	instance := instanceInterface.(vm.Instance)
	privateIpAddress := instance.PrivateIpAddress
	networkInterfaceId := instance.PrimaryNetworkInterface.NetworkInterface.NetworkInterfaceId

	s.ui.Message("Creating public-ip")
	s.InstanceSpecConfig.PublicIpId, err = createElasticIp(state)
	if err != nil {
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.ui.Message("Associating public-ip with instance")
	err = associatePublicIp(networkInterfaceId, s.InstanceSpecConfig.PublicIpId, privateIpAddress)
	if err != nil {
		s.ui.Error(err.Error())
		return multistep.ActionHalt
	}

	req_ := vpcApis.NewDescribeElasticIpRequest(Region, s.InstanceSpecConfig.PublicIpId)
	eip, err := VpcClient.DescribeElasticIp(req_)
	if err != nil || eip.Error.Code != FINE {
		s.ui.Error(fmt.Sprintf("[ERROR] Failed in getting eip,error:%v \n response:%v", err, eip))
		return multistep.ActionHalt
	}

	s.InstanceSpecConfig.PublicIpAddress = eip.Result.ElasticIp.ElasticIpAddress
	state.Put("eip", s.InstanceSpecConfig.PublicIpAddress)
	s.ui.Message(fmt.Sprintf(
		"Hi, we have created the instance, its name=%v , "+
			"its id=%v, "+
			"and its eip=%v  :) ", instance.InstanceName, s.InstanceSpecConfig.InstanceId, eip.Result.ElasticIp.ElasticIpAddress))
	return multistep.ActionContinue
}

// Delete created resources {instance,ip} on error
func (s *stepCreateJDCloudInstance) Cleanup(state multistep.StateBag) {

	if s.InstanceSpecConfig.PublicIpId != "" {

		req := vpcApis.NewDeleteElasticIpRequest(Region, s.InstanceSpecConfig.PublicIpId)

		_ = Retry(time.Minute, func() *RetryError {
			_, err := VpcClient.DeleteElasticIp(req)
			if err == nil {
				return nil
			}
			if connectionError(err) {
				return RetryableError(err)
			} else {
				return NonRetryableError(err)
			}
		})
	}

	if s.InstanceSpecConfig.InstanceId != "" {

		req := apis.NewDeleteInstanceRequest(Region, s.InstanceSpecConfig.InstanceId)
		_ = Retry(time.Minute, func() *RetryError {
			_, err := VmClient.DeleteInstance(req)
			if err == nil {
				return nil
			}
			if connectionError(err) {
				return RetryableError(err)
			} else {
				return NonRetryableError(err)
			}
		})
	}
}

func createElasticIp(state multistep.StateBag) (string, error) {

	generalConfig := state.Get("config").(Config)
	regionId := generalConfig.RegionId
	credential := core.NewCredentials(generalConfig.AccessKey, generalConfig.SecretKey)
	vpcclient := vpcClient.NewVpcClient(credential)

	req := vpcApis.NewCreateElasticIpsRequest(regionId, 1, &vpc.ElasticIpSpec{
		BandwidthMbps: 1,
		Provider:      "bgp",
	})

	resp, err := vpcclient.CreateElasticIps(req)

	if err != nil || resp.Error.Code != 0 {
		return "", fmt.Errorf("[ERROR] Failed in creating new publicIp, Error-%v, Response:%v", err, resp)
	}
	return resp.Result.ElasticIpIds[0], nil
}

func associatePublicIp(networkInterfaceId string, eipId string, privateIpAddress string) error {
	req := vpcApis.NewAssociateElasticIpRequest(Region, networkInterfaceId)
	req.ElasticIpId = &eipId
	req.PrivateIpAddress = &privateIpAddress
	resp, err := VpcClient.AssociateElasticIp(req)
	if err != nil || resp.Error.Code != FINE {
		return fmt.Errorf("[ERROR] Failed in associating publicIp, Error-%v, Response:%v", err, resp)
	}
	return nil
}
func instanceHost(state multistep.StateBag) (string, error) {
	return state.Get("eip").(string), nil
}

func InstanceStatusRefresher(id string, pending, target []string) (instance interface{}, err error) {

	stateConf := &StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    instanceStatusRefresher(id),
		Delay:      3 * time.Second,
		Timeout:    10 * time.Minute,
		MinTimeout: 1 * time.Second,
	}
	if instance, err = stateConf.WaitForState(); err != nil {
		return nil, fmt.Errorf("[ERROR] Failed in creating instance ,err message:%v", err)
	}
	return instance, nil
}

func instanceStatusRefresher(instanceId string) StateRefreshFunc {

	return func() (instance interface{}, status string, err error) {

		err = Retry(time.Minute, func() *RetryError {

			req := apis.NewDescribeInstanceRequest(Region, instanceId)
			resp, err := VmClient.DescribeInstance(req)

			if err == nil && resp.Error.Code == FINE {
				instance = resp.Result.Instance
				status = resp.Result.Instance.Status
				return nil
			}

			instance = nil
			status = ""
			if connectionError(err) {
				return RetryableError(err)
			} else {
				return NonRetryableError(err)
			}
		})
		return instance, status, err
	}
}

func connectionError(e error) bool {

	if e == nil {
		return false
	}
	ok, _ := regexp.MatchString(CONNECT_FAILED, e.Error())
	return ok
}
