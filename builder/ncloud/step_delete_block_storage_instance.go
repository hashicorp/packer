package ncloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepDeleteBlockStorageInstance struct {
	Conn                       *ncloud.Conn
	DeleteBlockStorageInstance func(blockStorageInstanceNo string) error
	Say                        func(message string)
	Error                      func(e error)
	Config                     *Config
}

func NewStepDeleteBlockStorageInstance(conn *ncloud.Conn, ui packer.Ui, config *Config) *StepDeleteBlockStorageInstance {
	var step = &StepDeleteBlockStorageInstance{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	step.DeleteBlockStorageInstance = step.deleteBlockStorageInstance

	return step
}

func (s *StepDeleteBlockStorageInstance) getBlockInstanceList(serverInstanceNo string) []string {
	reqParams := new(ncloud.RequestBlockStorageInstanceList)
	reqParams.ServerInstanceNo = serverInstanceNo

	blockStorageInstanceList, err := s.Conn.GetBlockStorageInstance(reqParams)
	if err != nil {
		return nil
	}

	if blockStorageInstanceList.TotalRows == 1 {
		return nil
	}

	var instanceList []string

	for _, blockStorageInstance := range blockStorageInstanceList.BlockStorageInstance {
		log.Println(blockStorageInstance)
		if blockStorageInstance.BlockStorageType.Code != "BASIC" {
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

	_, err := s.Conn.DeleteBlockStorageInstances(blockStorageInstanceList)
	if err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Block Storage Instance is deleted. Block Storage InstanceNo is %s", blockStorageInstanceList))

	if err := waiterDetachedBlockStorageInstance(s.Conn, serverInstanceNo, time.Minute); err != nil {
		return errors.New("TIMEOUT : Block Storage instance status is not deattached")
	}

	return nil
}

func (s *StepDeleteBlockStorageInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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
