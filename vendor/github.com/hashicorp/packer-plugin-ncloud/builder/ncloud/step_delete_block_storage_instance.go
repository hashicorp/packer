package ncloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepDeleteBlockStorage struct {
	Conn               *NcloudAPIClient
	DeleteBlockStorage func(blockStorageNo string) error
	Say                func(message string)
	Error              func(e error)
	Config             *Config
}

func NewStepDeleteBlockStorage(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepDeleteBlockStorage {
	var step = &StepDeleteBlockStorage{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.DeleteBlockStorage = step.deleteVpcBlockStorage
	} else {
		step.DeleteBlockStorage = step.deleteClassicBlockStorage
	}

	return step
}

func (s *StepDeleteBlockStorage) getClassicBlockList(serverInstanceNo string) []*string {
	reqParams := &server.GetBlockStorageInstanceListRequest{
		ServerInstanceNo: &serverInstanceNo,
	}

	blockStorageInstanceList, err := s.Conn.server.V2Api.GetBlockStorageInstanceList(reqParams)
	if err != nil {
		return nil
	}

	if *blockStorageInstanceList.TotalRows == 1 {
		return nil
	}

	var instanceList []*string

	for _, blockStorageInstance := range blockStorageInstanceList.BlockStorageInstanceList {
		log.Println(blockStorageInstance)
		if *blockStorageInstance.BlockStorageType.Code != "BASIC" {
			instanceList = append(instanceList, blockStorageInstance.BlockStorageInstanceNo)
		}
	}

	return instanceList
}

func (s *StepDeleteBlockStorage) getVpcBlockList(serverInstanceNo string) []*string {
	reqParams := &vserver.GetBlockStorageInstanceListRequest{
		ServerInstanceNo: &serverInstanceNo,
	}

	blockStorageInstanceList, err := s.Conn.vserver.V2Api.GetBlockStorageInstanceList(reqParams)
	if err != nil {
		return nil
	}

	if *blockStorageInstanceList.TotalRows == 1 {
		return nil
	}

	var instanceList []*string

	for _, blockStorageInstance := range blockStorageInstanceList.BlockStorageInstanceList {
		log.Println(blockStorageInstance)
		if *blockStorageInstance.BlockStorageType.Code != "BASIC" {
			instanceList = append(instanceList, blockStorageInstance.BlockStorageInstanceNo)
		}
	}

	return instanceList
}

func (s *StepDeleteBlockStorage) deleteClassicBlockStorage(serverInstanceNo string) error {
	blockStorageInstanceList := s.getClassicBlockList(serverInstanceNo)
	if len(blockStorageInstanceList) == 0 {
		return nil
	}
	reqParams := server.DeleteBlockStorageInstancesRequest{
		BlockStorageInstanceNoList: blockStorageInstanceList,
	}
	_, err := s.Conn.server.V2Api.DeleteBlockStorageInstances(&reqParams)
	if err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Block Storage Instance is deleted. Block Storage Instance List is %v", blockStorageInstanceList))

	if err := waiterClassicDetachedBlockStorage(s.Conn, serverInstanceNo, time.Minute); err != nil {
		return errors.New("TIMEOUT : Block Storage instance status is not deattached")
	}

	return nil
}

func (s *StepDeleteBlockStorage) deleteVpcBlockStorage(serverInstanceNo string) error {
	blockStorageInstanceList := s.getVpcBlockList(serverInstanceNo)
	if len(blockStorageInstanceList) == 0 {
		return nil
	}
	reqParams := vserver.DeleteBlockStorageInstancesRequest{
		BlockStorageInstanceNoList: blockStorageInstanceList,
	}
	_, err := s.Conn.vserver.V2Api.DeleteBlockStorageInstances(&reqParams)
	if err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Block Storage Instance is deleted. Block Storage Instance List is %v", blockStorageInstanceList))

	if err := waiterVpcDetachedBlockStorage(s.Conn, serverInstanceNo, time.Minute); err != nil {
		return errors.New("TIMEOUT : Block Storage instance status is not deattached")
	}

	return nil
}

func (s *StepDeleteBlockStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.Config.BlockStorageSize == 0 {
		return processStepResult(nil, s.Error, state)
	}

	s.Say("Delete Block Storage Instance")

	var serverInstanceNo = state.Get("instance_no").(string)

	err := s.DeleteBlockStorage(serverInstanceNo)

	return processStepResult(err, s.Error, state)
}

func (*StepDeleteBlockStorage) Cleanup(multistep.StateBag) {
}
