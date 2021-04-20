package ncloud

import (
	"context"
	"errors"
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
	BlockStorageStatusAttached = "ATTAC"
	BlockStorageStatusDetached = "DETAC"
)

// StepCreateBlockStorageInstance struct is for making extra block storage
type StepCreateBlockStorage struct {
	Conn                     *NcloudAPIClient
	CreateBlockStorage       func(serverInstanceNo string) (*string, error)
	DeleteBlockStorage       func(blockStorageInstanceNo string) error
	WaiterBlockStorageStatus func(conn *NcloudAPIClient, blockStorageInstanceNo *string, status string, timeout time.Duration) error
	Say                      func(message string)
	Error                    func(e error)
	Config                   *Config
}

// NewStepCreateBlockStorageInstance make StepCreateBlockStorage struct to make extra block storage
func NewStepCreateBlockStorage(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepCreateBlockStorage {
	var step = &StepCreateBlockStorage{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.CreateBlockStorage = step.createVpcBlockStorage
		step.DeleteBlockStorage = step.deleteVpcBlockStorage
		step.WaiterBlockStorageStatus = waiterVpcBlockStorageStatus
	} else {
		step.CreateBlockStorage = step.createClassicBlockStorage
		step.DeleteBlockStorage = step.deleteClassicBlockStorage
		step.WaiterBlockStorageStatus = waiterClassicBlockStorageStatus
	}

	return step
}

func (s *StepCreateBlockStorage) createClassicBlockStorage(serverInstanceNo string) (*string, error) {
	reqParams := &server.CreateBlockStorageInstanceRequest{
		BlockStorageSize: ncloud.Int64(int64(s.Config.BlockStorageSize)),
		ServerInstanceNo: &serverInstanceNo,
	}

	resp, err := s.Conn.server.V2Api.CreateBlockStorageInstance(reqParams)
	if err != nil {
		return nil, err
	}

	blockStorageInstance := resp.BlockStorageInstanceList[0]
	log.Println("Block Storage Instance information : ", blockStorageInstance.BlockStorageInstanceNo)

	if err := waiterClassicBlockStorageStatus(s.Conn, blockStorageInstance.BlockStorageInstanceNo, BlockStorageStatusAttached, 10*time.Minute); err != nil {
		return nil, errors.New("TIMEOUT : Block Storage instance status is not attached")
	}

	return blockStorageInstance.BlockStorageInstanceNo, nil
}

func (s *StepCreateBlockStorage) createVpcBlockStorage(serverInstanceNo string) (*string, error) {
	reqParams := &vserver.CreateBlockStorageInstanceRequest{
		RegionCode:                     &s.Config.RegionCode,
		BlockStorageSize:               ncloud.Int32(int32(s.Config.BlockStorageSize)),
		BlockStorageDescription:        nil,
		ServerInstanceNo:               &serverInstanceNo,
		BlockStorageSnapshotInstanceNo: nil,
		ZoneCode:                       nil,
	}

	resp, err := s.Conn.vserver.V2Api.CreateBlockStorageInstance(reqParams)
	if err != nil {
		return nil, err
	}

	blockStorageInstance := resp.BlockStorageInstanceList[0]
	log.Println("Block Storage Instance information : ", blockStorageInstance.BlockStorageInstanceNo)

	if err := s.WaiterBlockStorageStatus(s.Conn, blockStorageInstance.BlockStorageInstanceNo, BlockStorageStatusAttached, 10*time.Minute); err != nil {
		return nil, errors.New("TIMEOUT : Block Storage instance status is not attached")
	}

	return blockStorageInstance.BlockStorageInstanceNo, nil
}

func (s *StepCreateBlockStorage) deleteClassicBlockStorage(blockStorageInstanceNo string) error {
	reqParams := &server.DeleteBlockStorageInstancesRequest{
		BlockStorageInstanceNoList: []*string{&blockStorageInstanceNo},
	}

	_, err := s.Conn.server.V2Api.DeleteBlockStorageInstances(reqParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *StepCreateBlockStorage) deleteVpcBlockStorage(blockStorageInstanceNo string) error {
	reqParams := &vserver.DeleteBlockStorageInstancesRequest{
		BlockStorageInstanceNoList: []*string{&blockStorageInstanceNo},
	}

	_, err := s.Conn.vserver.V2Api.DeleteBlockStorageInstances(reqParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *StepCreateBlockStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.Config.BlockStorageSize == 0 {
		return processStepResult(nil, s.Error, state)
	}

	s.Say("Create extra block storage instance")

	serverInstanceNo := state.Get("instance_no").(string)

	blockStorageInstanceNo, err := s.CreateBlockStorage(serverInstanceNo)
	if err == nil {
		state.Put("block_storage_instance_no", *blockStorageInstanceNo)
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreateBlockStorage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	if s.Config.BlockStorageSize == 0 {
		return
	}

	if blockStorageInstanceNo, ok := state.GetOk("block_storage_instance_no"); ok {
		s.Say("Clean up Block Storage Instance")
		err := s.DeleteBlockStorage(blockStorageInstanceNo.(string))
		if err != nil {
			s.Error(err)
			return
		}

		s.Say(fmt.Sprintf("Block Storage Instance is deleted. Block Storage InstanceNo is %s", blockStorageInstanceNo.(string)))

		if err := s.WaiterBlockStorageStatus(s.Conn, blockStorageInstanceNo.(*string), BlockStorageStatusDetached, time.Minute); err != nil {
			s.Say("TIMEOUT : Block Storage instance status is not deattached")
		}
	}
}
