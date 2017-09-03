package ecs

import (
	"errors"
	"fmt"
	"time"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepConfigAlicloudVSwitch struct {
	VSwitchId   string
	ZoneId      string
	isCreate    bool
	CidrBlock   string
	VSwitchName string
}

func (s *stepConfigAlicloudVSwitch) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	vpcId := state.Get("vpcid").(string)
	config := state.Get("config").(Config)

	if len(s.VSwitchId) != 0 {
		vswitchs, _, err := client.DescribeVSwitches(&ecs.DescribeVSwitchesArgs{
			VpcId:     vpcId,
			VSwitchId: s.VSwitchId,
			ZoneId:    s.ZoneId,
		})
		if err != nil {
			ui.Say(fmt.Sprintf("Failed querying vswitch: %s", err))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if len(vswitchs) > 0 {
			vswitch := vswitchs[0]
			state.Put("vswitchid", vswitch.VSwitchId)
			s.isCreate = false
			return multistep.ActionContinue
		}
		s.isCreate = false
		message := fmt.Sprintf("The specified vswitch {%s} doesn't exist.", s.VSwitchId)
		state.Put("error", errors.New(message))
		ui.Say(message)
		return multistep.ActionHalt

	}
	if s.ZoneId == "" {

		zones, err := client.DescribeZones(common.Region(config.AlicloudRegion))
		if err != nil {
			ui.Say(fmt.Sprintf("Query for available zones failed: %s", err))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		var instanceTypes []string
		for _, zone := range zones {
			isVSwitchSupported := false
			for _, resourceType := range zone.AvailableResourceCreation.ResourceTypes {
				if resourceType == ecs.ResourceTypeVSwitch {
					isVSwitchSupported = true
				}
			}
			if isVSwitchSupported {
				for _, instanceType := range zone.AvailableInstanceTypes.InstanceTypes {
					if instanceType == config.InstanceType {
						s.ZoneId = zone.ZoneId
						break
					}
					instanceTypes = append(instanceTypes, instanceType)
				}
			}
		}

		if s.ZoneId == "" {
			if len(instanceTypes) > 0 {
				ui.Say(fmt.Sprintf("The instance type %s isn't available in this region."+
					"\n You can either change the instance to one of following: %v \n"+
					"or choose another region.", config.InstanceType, instanceTypes))

				state.Put("error", fmt.Errorf("The instance type %s isn't available in this region."+
					"\n You can either change the instance to one of following: %v \n"+
					"or choose another region.", config.InstanceType, instanceTypes))
				return multistep.ActionHalt
			} else {
				ui.Say(fmt.Sprintf("The instance type %s isn't available in this region."+
					"\n You can change to other regions.", config.InstanceType))

				state.Put("error", fmt.Errorf("The instance type %s isn't available in this region."+
					"\n You can change to other regions.", config.InstanceType))
				return multistep.ActionHalt
			}
		}
	}
	if config.CidrBlock == "" {
		s.CidrBlock = "172.16.0.0/24" //use the default CirdBlock
	}
	ui.Say("Creating vswitch...")
	vswitchId, err := client.CreateVSwitch(&ecs.CreateVSwitchArgs{
		CidrBlock:   s.CidrBlock,
		ZoneId:      s.ZoneId,
		VpcId:       vpcId,
		VSwitchName: s.VSwitchName,
	})
	if err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Create vswitch failed %v", err))
		return multistep.ActionHalt
	}
	if err := client.WaitForVSwitchAvailable(vpcId, s.VSwitchId, ALICLOUD_DEFAULT_TIMEOUT); err != nil {
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Timeout waiting for vswitch to become available: %v", err))
		return multistep.ActionHalt
	}
	state.Put("vswitchid", vswitchId)
	s.isCreate = true
	s.VSwitchId = vswitchId
	return multistep.ActionContinue
}

func (s *stepConfigAlicloudVSwitch) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	message(state, "vSwitch")
	timeoutPoint := time.Now().Add(10 * time.Second)
	for {
		if err := client.DeleteVSwitch(s.VSwitchId); err != nil {
			e, _ := err.(*common.Error)
			if (e.Code == "IncorrectVSwitchStatus" || e.Code == "DependencyViolation" ||
				e.Code == "DependencyViolation.HaVip" ||
				e.Code == "IncorretRouteEntryStatus") && time.Now().Before(timeoutPoint) {
				time.Sleep(1 * time.Second)
				continue
			}
			ui.Error(fmt.Sprintf("Error deleting vswitch, it may still be around: %s", err))
			return
		}
		break
	}
}
