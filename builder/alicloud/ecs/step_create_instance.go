package ecs

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateAlicloudInstance struct {
	IOOptimized             bool
	InstanceType            string
	UserData                string
	UserDataFile            string
	instanceId              string
	RegionId                string
	InternetChargeType      string
	InternetMaxBandwidthOut int
	InstnaceName            string
	ZoneId                  string
	instance                *ecs.InstanceAttributesType
}

func (s *stepCreateAlicloudInstance) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)
	source_image := state.Get("source_image").(*ecs.ImageType)
	network_type := state.Get("networktype").(InstanceNetWork)
	securityGroupId := state.Get("securitygroupid").(string)
	var instanceId string
	var err error

	ioOptimized := ecs.IoOptimizedNone
	if s.IOOptimized {
		ioOptimized = ecs.IoOptimizedOptimized
	}
	password := config.Comm.SSHPassword
	if password == "" && config.Comm.WinRMPassword != "" {
		password = config.Comm.WinRMPassword
	}
	ui.Say("Creating instance.")
	if network_type == VpcNet {
		userData, err := s.getUserData(state)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		vswitchId := state.Get("vswitchid").(string)
		instanceId, err = client.CreateInstance(&ecs.CreateInstanceArgs{
			RegionId:                common.Region(s.RegionId),
			ImageId:                 source_image.ImageId,
			InstanceType:            s.InstanceType,
			InternetChargeType:      common.InternetChargeType(s.InternetChargeType), //"PayByTraffic",
			InternetMaxBandwidthOut: s.InternetMaxBandwidthOut,
			UserData:                userData,
			IoOptimized:             ioOptimized,
			VSwitchId:               vswitchId,
			SecurityGroupId:         securityGroupId,
			InstanceName:            s.InstnaceName,
			Password:                password,
			ZoneId:                  s.ZoneId,
			DataDisk:                diskDeviceToDiskType(config.AlicloudImageConfig.ECSImagesDiskMappings),
		})
		if err != nil {
			err := fmt.Errorf("Error creating instance: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		if s.InstanceType == "" {
			s.InstanceType = "PayByTraffic"
		}
		if s.InternetMaxBandwidthOut == 0 {
			s.InternetMaxBandwidthOut = 5
		}
		instanceId, err = client.CreateInstance(&ecs.CreateInstanceArgs{
			RegionId:                common.Region(s.RegionId),
			ImageId:                 source_image.ImageId,
			InstanceType:            s.InstanceType,
			InternetChargeType:      common.InternetChargeType(s.InternetChargeType), //"PayByTraffic",
			InternetMaxBandwidthOut: s.InternetMaxBandwidthOut,
			IoOptimized:             ioOptimized,
			SecurityGroupId:         securityGroupId,
			InstanceName:            s.InstnaceName,
			Password:                password,
			ZoneId:                  s.ZoneId,
			DataDisk:                diskDeviceToDiskType(config.AlicloudImageConfig.ECSImagesDiskMappings),
		})
		if err != nil {
			err := fmt.Errorf("Error creating instance: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}
	err = client.WaitForInstance(instanceId, ecs.Stopped, ALICLOUD_DEFAULT_TIMEOUT)
	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	instance, err := client.DescribeInstanceAttribute(instanceId)
	if err != nil {
		ui.Say(err.Error())
		return multistep.ActionHalt
	}
	s.instance = instance
	state.Put("instance", instance)

	return multistep.ActionContinue
}

func (s *stepCreateAlicloudInstance) Cleanup(state multistep.StateBag) {
	if s.instance == nil {
		return
	}
	message(state, "instance")
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	err := client.DeleteInstance(s.instance.InstanceId)
	if err != nil {
		ui.Say(fmt.Sprintf("Failed to clean up instance %s: %v", s.instance.InstanceId, err.Error()))
	}

}

func (s *stepCreateAlicloudInstance) getUserData(state multistep.StateBag) (string, error) {
	userData := s.UserData
	if s.UserDataFile != "" {
		data, err := ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			return "", err
		}
		userData = string(data)
	}
	log.Printf(userData)
	return userData, nil

}

func diskDeviceToDiskType(diskDevices []AlicloudDiskDevice) []ecs.DataDiskType {
	result := make([]ecs.DataDiskType, len(diskDevices))
	for _, diskDevice := range diskDevices {
		result = append(result, ecs.DataDiskType{
			DiskName:           diskDevice.DiskName,
			Category:           ecs.DiskCategory(diskDevice.DiskCategory),
			Size:               diskDevice.DiskSize,
			SnapshotId:         diskDevice.SnapshotId,
			Description:        diskDevice.Description,
			DeleteWithInstance: diskDevice.DeleteWithInstance,
			Device:             diskDevice.Device,
		})
	}
	return result
}
