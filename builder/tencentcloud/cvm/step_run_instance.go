package cvm

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
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
	Tags                     map[string]string
	DataDisks                []tencentCloudDataDisk
}

func (s *stepRunInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("cvm_client").(*cvm.Client)

	config := state.Get("config").(*Config)
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
		return Halt(state, err, "Failed to get user_data")
	}

	Say(state, "Trying to create a new instance", "")

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
	// TODO: Add check for system disk size, it should be larger than image system disk size.
	req.SystemDisk = &cvm.SystemDisk{
		DiskType: &s.DiskType,
		DiskSize: &s.DiskSize,
	}
	// System disk snapshot is mandatory, so if there are additional data disks,
	// length will be larger than 1.
	if source_image.SnapshotSet != nil && len(source_image.SnapshotSet) > 1 {
		Message(state, "Use source image snapshot data disks, ignore user data disk settings", "")
		var dataDisks []*cvm.DataDisk
		for _, snapshot := range source_image.SnapshotSet {
			if *snapshot.DiskUsage == "DATA_DISK" {
				var dataDisk cvm.DataDisk
				// FIXME: Currently we have no way to get original disk type
				// from data disk snapshots, and we don't allow user to overwrite
				// snapshot settings, and we cannot guarantee a certain hard-coded type
				// is not sold out, so here we use system disk type as a workaround.
				//
				// Eventually, we need to allow user to overwrite snapshot disk
				// settings.
				dataDisk.DiskType = &s.DiskType
				dataDisk.DiskSize = snapshot.DiskSize
				dataDisk.SnapshotId = snapshot.SnapshotId
				dataDisks = append(dataDisks, &dataDisk)
			}
		}
		req.DataDisks = dataDisks
	} else {
		var dataDisks []*cvm.DataDisk
		for _, disk := range s.DataDisks {
			var dataDisk cvm.DataDisk
			dataDisk.DiskType = &disk.DiskType
			dataDisk.DiskSize = &disk.DiskSize
			if disk.SnapshotId != "" {
				dataDisk.SnapshotId = &disk.SnapshotId
			}
			dataDisks = append(dataDisks, &dataDisk)
		}
		req.DataDisks = dataDisks
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
	var tags []*cvm.Tag
	for k, v := range s.Tags {
		tags = append(tags, &cvm.Tag{
			Key:   &k,
			Value: &v,
		})
	}
	resourceType := "instance"
	if len(tags) > 0 {
		req.TagSpecification = []*cvm.TagSpecification{
			{
				ResourceType: &resourceType,
				Tags:         tags,
			},
		}
	}

	var resp *cvm.RunInstancesResponse
	err = Retry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = client.RunInstances(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to run instance")
	}

	if len(resp.Response.InstanceIdSet) != 1 {
		return Halt(state, fmt.Errorf("No instance return"), "Failed to run instance")
	}

	s.instanceId = *resp.Response.InstanceIdSet[0]
	Message(state, "Waiting for instance ready", "")

	err = WaitForInstance(ctx, client, s.instanceId, "RUNNING", 1800)
	if err != nil {
		return Halt(state, err, "Failed to wait for instance ready")
	}

	describeReq := cvm.NewDescribeInstancesRequest()
	describeReq.InstanceIds = []*string{&s.instanceId}
	var describeResp *cvm.DescribeInstancesResponse
	err = Retry(ctx, func(ctx context.Context) error {
		var e error
		describeResp, e = client.DescribeInstances(describeReq)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to wait for instance ready")
	}

	state.Put("instance", describeResp.Response.InstanceSet[0])
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", s.instanceId)
	Message(state, s.instanceId, "Instance created")

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
	log.Printf(fmt.Sprintf("[DEBUG]getUserData: user_data: %s", userData))

	return userData, nil
}

func (s *stepRunInstance) Cleanup(state multistep.StateBag) {
	if s.instanceId == "" {
		return
	}

	ctx := context.TODO()
	client := state.Get("cvm_client").(*cvm.Client)

	SayClean(state, "instance")

	req := cvm.NewTerminateInstancesRequest()
	req.InstanceIds = []*string{&s.instanceId}
	err := Retry(ctx, func(ctx context.Context) error {
		_, e := client.TerminateInstances(req)
		return e
	})
	if err != nil {
		Error(state, err, fmt.Sprintf("Failed to terminate instance(%s), please delete it manually", s.instanceId))
	}
}
