package common

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/outscale/osc-go/oapi"

	retry "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const (
	RunSourceVmBSUExpectedRootDevice = "ebs"
)

type StepRunSourceVm struct {
	AssociatePublicIpAddress    bool
	BlockDevices                BlockDevices
	Comm                        *communicator.Config
	Ctx                         interpolate.Context
	Debug                       bool
	BsuOptimized                bool
	EnableT2Unlimited           bool
	ExpectedRootDevice          string
	IamVmProfile                string
	VmInitiatedShutdownBehavior string
	VmType                      string
	IsRestricted                bool
	SourceOMI                   string
	Tags                        TagMap
	UserData                    string
	UserDataFile                string
	VolumeTags                  TagMap

	vmId string
}

func (s *StepRunSourceVm) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)

	securityGroupIds := state.Get("securityGroupIds").([]string)
	ui := state.Get("ui").(packer.Ui)

	userData := s.UserData
	if s.UserDataFile != "" {
		contents, err := ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			state.Put("error", fmt.Errorf("Problem reading user data file: %s", err))
			return multistep.ActionHalt
		}

		userData = string(contents)
	}

	// Test if it is encoded already, and if not, encode it
	if _, err := base64.StdEncoding.DecodeString(userData); err != nil {
		log.Printf("[DEBUG] base64 encoding user data...")
		userData = base64.StdEncoding.EncodeToString([]byte(userData))
	}

	ui.Say("Launching a source OUTSCALE vm...")
	image, ok := state.Get("source_image").(oapi.Image)
	if !ok {
		state.Put("error", fmt.Errorf("source_image type assertion failed"))
		return multistep.ActionHalt
	}
	s.SourceOMI = image.ImageId

	if s.ExpectedRootDevice != "" && image.RootDeviceType != s.ExpectedRootDevice {
		state.Put("error", fmt.Errorf(
			"The provided source OMI has an invalid root device type.\n"+
				"Expected '%s', got '%s'.",
			s.ExpectedRootDevice, image.RootDeviceType))
		return multistep.ActionHalt
	}

	var vmId string

	ui.Say("Adding tags to source vm")
	if _, exists := s.Tags["Name"]; !exists {
		s.Tags["Name"] = "Packer Builder"
	}

	oapiTags, err := s.Tags.OAPITags(s.Ctx, oapiconn.GetConfig().Region, state)
	if err != nil {
		err := fmt.Errorf("Error tagging source vm: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	volTags, err := s.VolumeTags.OAPITags(s.Ctx, oapiconn.GetConfig().Region, state)
	if err != nil {
		err := fmt.Errorf("Error tagging volumes: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	subregion := state.Get("subregion_name").(string)
	runOpts := oapi.CreateVmsRequest{
		ImageId:             s.SourceOMI,
		VmType:              s.VmType,
		UserData:            userData,
		MaxVmsCount:         1,
		MinVmsCount:         1,
		Placement:           oapi.Placement{SubregionName: subregion},
		BsuOptimized:        s.BsuOptimized,
		BlockDeviceMappings: s.BlockDevices.BuildLaunchDevices(),
		//IamVmProfile:        oapi.IamVmProfileSpecification{Name: &s.IamVmProfile},
	}

	// if s.EnableT2Unlimited {
	// 	creditOption := "unlimited"
	// 	runOpts.CreditSpecification = &oapi.CreditSpecificationRequest{CpuCredits: &creditOption}
	// }

	// Collect tags for tagging on resource creation
	//	var tagSpecs []oapi.ResourceTag

	// if len(oapiTags) > 0 {
	// 	runTags := &oapi.ResourceTag{
	// 		ResourceType: aws.String("vm"),
	// 		Tags:         oapiTags,
	// 	}

	// 	tagSpecs = append(tagSpecs, runTags)
	// }

	// if len(volTags) > 0 {
	// 	runVolTags := &oapi.TagSpecification{
	// 		ResourceType: aws.String("volume"),
	// 		Tags:         volTags,
	// 	}

	// 	tagSpecs = append(tagSpecs, runVolTags)
	// }

	// // If our region supports it, set tag specifications
	// if len(tagSpecs) > 0 && !s.IsRestricted {
	// 	runOpts.SetTagSpecifications(tagSpecs)
	// 	oapiTags.Report(ui)
	// 	volTags.Report(ui)
	// }

	if s.Comm.SSHKeyPairName != "" {
		runOpts.KeypairName = s.Comm.SSHKeyPairName
	}

	subnetId := state.Get("subnet_id").(string)

	if subnetId != "" && s.AssociatePublicIpAddress {
		runOpts.Nics = []oapi.NicForVmCreation{
			{
				DeviceNumber: 0,
				//AssociatePublicIpAddress: s.AssociatePublicIpAddress,
				SubnetId:           subnetId,
				SecurityGroupIds:   securityGroupIds,
				DeleteOnVmDeletion: true,
			},
		}
	} else {
		runOpts.SubnetId = subnetId
		runOpts.SecurityGroupIds = securityGroupIds
	}

	if s.ExpectedRootDevice == "bsu" {
		runOpts.VmInitiatedShutdownBehavior = s.VmInitiatedShutdownBehavior
	}

	runResp, err := oapiconn.POST_CreateVms(runOpts)
	if err != nil {
		err := fmt.Errorf("Error launching source vm: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	vmId = runResp.OK.Vms[0].VmId
	volumeId := runResp.OK.Vms[0].BlockDeviceMappings[0].Bsu.VolumeId

	// Set the vm ID so that the cleanup works properly
	s.vmId = vmId

	ui.Message(fmt.Sprintf("Vm ID: %s", vmId))
	ui.Say(fmt.Sprintf("Waiting for vm (%v) to become ready...", vmId))

	request := oapi.ReadVmsRequest{
		Filters: oapi.FiltersVm{
			VmIds: []string{vmId},
		},
	}
	if err := waitUntilForVmRunning(oapiconn, vmId); err != nil {
		err := fmt.Errorf("Error waiting for vm (%s) to become ready: %s", vmId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	//Set Vm tags and vollume tags
	if len(oapiTags) > 0 {
		if err := CreateTags(oapiconn, s.vmId, ui, oapiTags); err != nil {
			err := fmt.Errorf("Error creating tags for vm (%s): %s", s.vmId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if len(volTags) > 0 {
		if err := CreateTags(oapiconn, volumeId, ui, volTags); err != nil {
			err := fmt.Errorf("Error creating tags for volume (%s): %s", volumeId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	//TODO: LinkPublicIp i

	resp, err := oapiconn.POST_ReadVms(request)

	r := resp.OK

	if err != nil || len(r.Vms) == 0 {
		err := fmt.Errorf("Error finding source vm.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	vm := r.Vms[0]

	if s.Debug {
		if vm.PublicDnsName != "" {
			ui.Message(fmt.Sprintf("Public DNS: %s", vm.PublicDnsName))
		}

		if vm.PublicIp != "" {
			ui.Message(fmt.Sprintf("Public IP: %s", vm.PublicIp))
		}

		if vm.PrivateIp != "" {
			ui.Message(fmt.Sprintf("Private IP: %s", vm.PublicIp))
		}
	}

	state.Put("vm", vm)

	// If we're in a region that doesn't support tagging on vm creation,
	// do that now.

	if s.IsRestricted {
		oapiTags.Report(ui)
		// Retry creating tags for about 2.5 minutes
		err = retry.Retry(0.2, 30, 11, func(_ uint) (bool, error) {
			_, err := oapiconn.POST_CreateTags(oapi.CreateTagsRequest{
				Tags:        oapiTags,
				ResourceIds: []string{vmId},
			})
			if err == nil {
				return true, nil
			}
			//TODO: improve error
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidVmID.NotFound" {
					return false, nil
				}
			}
			return true, err
		})

		if err != nil {
			err := fmt.Errorf("Error tagging source vm: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Now tag volumes

		volumeIds := make([]string, 0)
		for _, v := range vm.BlockDeviceMappings {
			if bsu := v.Bsu; !reflect.DeepEqual(bsu, oapi.BsuCreated{}) {
				volumeIds = append(volumeIds, bsu.VolumeId)
			}
		}

		if len(volumeIds) > 0 && s.VolumeTags.IsSet() {
			ui.Say("Adding tags to source BSU Volumes")

			volumeTags, err := s.VolumeTags.OAPITags(s.Ctx, oapiconn.GetConfig().Region, state)
			if err != nil {
				err := fmt.Errorf("Error tagging source BSU Volumes on %s: %s", vm.VmId, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			volumeTags.Report(ui)

			_, err = oapiconn.POST_CreateTags(oapi.CreateTagsRequest{
				ResourceIds: volumeIds,
				Tags:        volumeTags,
			})

			if err != nil {
				err := fmt.Errorf("Error tagging source BSU Volumes on %s: %s", vm.VmId, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepRunSourceVm) Cleanup(state multistep.StateBag) {

	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)

	// Terminate the source vm if it exists
	if s.vmId != "" {
		ui.Say("Terminating the source OUTSCALE vm...")
		if _, err := oapiconn.POST_DeleteVms(oapi.DeleteVmsRequest{VmIds: []string{s.vmId}}); err != nil {
			ui.Error(fmt.Sprintf("Error terminating vm, may still be around: %s", err))
			return
		}

		if err := waitUntilVmDeleted(oapiconn, s.vmId); err != nil {
			ui.Error(err.Error())
		}
	}
}
