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

// StepCreateBlockStorageInstance struct is for making extra block storage
type StepCreateBlockStorageInstance struct {
	Conn                       *ncloud.Conn
	CreateBlockStorageInstance func(serverInstanceNo string) (string, error)
	Say                        func(message string)
	Error                      func(e error)
	Config                     *Config
}

// NewStepCreateBlockStorageInstance make StepCreateBlockStorage struct to make extra block storage
func NewStepCreateBlockStorageInstance(conn *ncloud.Conn, ui packer.Ui, config *Config) *StepCreateBlockStorageInstance {
	var step = &StepCreateBlockStorageInstance{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	step.CreateBlockStorageInstance = step.createBlockStorageInstance

	return step
}

func (s *StepCreateBlockStorageInstance) createBlockStorageInstance(serverInstanceNo string) (string, error) {

	reqParams := new(ncloud.RequestBlockStorageInstance)
	reqParams.BlockStorageSize = s.Config.BlockStorageSize
	reqParams.ServerInstanceNo = serverInstanceNo

	blockStorageInstanceList, err := s.Conn.CreateBlockStorageInstance(reqParams)
	if err != nil {
		return "", err
	}

	log.Println("Block Storage Instance information : ", blockStorageInstanceList.BlockStorageInstance[0])

	if err := waiterBlockStorageInstanceStatus(s.Conn, blockStorageInstanceList.BlockStorageInstance[0].BlockStorageInstanceNo, "ATTAC", 10*time.Minute); err != nil {
		return "", errors.New("TIMEOUT : Block Storage instance status is not attached")
	}

	return blockStorageInstanceList.BlockStorageInstance[0].BlockStorageInstanceNo, nil
}

func (s *StepCreateBlockStorageInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	if s.Config.BlockStorageSize == 0 {
		return processStepResult(nil, s.Error, state)
	}

	s.Say("Create extra block storage instance")

	serverInstanceNo := state.Get("InstanceNo").(string)

	blockStorageInstanceNo, err := s.CreateBlockStorageInstance(serverInstanceNo)
	if err == nil {
		state.Put("BlockStorageInstanceNo", blockStorageInstanceNo)
	}

	return processStepResult(err, s.Error, state)
}

func (s *StepCreateBlockStorageInstance) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	if s.Config.BlockStorageSize == 0 {
		return
	}

	if blockStorageInstanceNo, ok := state.GetOk("BlockStorageInstanceNo"); ok {
		s.Say("Clean up Block Storage Instance")
		no := blockStorageInstanceNo.(string)
		blockStorageInstanceList, err := s.Conn.DeleteBlockStorageInstances([]string{no})
		if err != nil {
			return
		}

		s.Say(fmt.Sprintf("Block Storage Instance is deleted. Block Storage InstanceNo is %s", no))
		log.Println("Block Storage Instance information : ", blockStorageInstanceList.BlockStorageInstance[0])

		if err := waiterBlockStorageInstanceStatus(s.Conn, no, "DETAC", time.Minute); err != nil {
			s.Say("TIMEOUT : Block Storage instance status is not deattached")
		}
	}
}
