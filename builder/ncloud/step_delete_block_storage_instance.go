package ncloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepDeleteBlockStorageInstance struct {
	Conn                       *NcloudAPIClient
	DeleteBlockStorageInstance func(blockStorageInstanceNo string) error
	Say                        func(message string)
	Error                      func(e error)
	Config                     *Config
}

func NewStepDeleteBlockStorageInstance(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepDeleteBlockStorageInstance {
	var step = &StepDeleteBlockStorageInstance{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	step.DeleteBlockStorageInstance = step.deleteBlockStorageInstance

	return step
}

func (s *StepDeleteBlockStorageInstance) getBlockInstanceList(serverInstanceNo string) []*string {
	reqParams := new(server.GetBlockStorageInstanceListRequest)
	reqParams.ServerInstanceNo = &serverInstanceNo

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

func (s *StepDeleteBlockStorageInstance) deleteBlockStorageInstance(serverInstanceNo string) error {
	blockStorageInstanceList := s.getBlockInstanceList(serverInstanceNo)
	if blockStorageInstanceList == nil || len(blockStorageInstanceList) == 0 {
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

	if err := waiterDetachedBlockStorageInstance(s.Conn, serverInstanceNo, time.Minute); err != nil {
		return errors.New("TIMEOUT : Block Storage instance status is not deattached")
	}

	return nil
}

func (s *StepDeleteBlockStorageInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.Config.BlockStorageSize == 0 {
		return processStepResult(nil, s.Error, state)
	}

	s.Say("Delete Block Storage Instance")

	var serverInstanceNo = state.Get("InstanceNo").(string)

	err := s.DeleteBlockStorageInstance(serverInstanceNo)

	return processStepResult(err, s.Error, state)
}

func (*StepDeleteBlockStorageInstance) Cleanup(multistep.StateBag) {
}
