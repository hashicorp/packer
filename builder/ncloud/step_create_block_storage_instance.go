package ncloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepCreateBlockStorageInstance struct is for making extra block storage
type StepCreateBlockStorageInstance struct {
	Conn                       *NcloudAPIClient
	CreateBlockStorageInstance func(serverInstanceNo string) (*string, error)
	Say                        func(message string)
	Error                      func(e error)
	Config                     *Config
}

// NewStepCreateBlockStorageInstance make StepCreateBlockStorage struct to make extra block storage
func NewStepCreateBlockStorageInstance(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepCreateBlockStorageInstance {
	var step = &StepCreateBlockStorageInstance{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	step.CreateBlockStorageInstance = step.createBlockStorageInstance

	return step
}

func (s *StepCreateBlockStorageInstance) createBlockStorageInstance(serverInstanceNo string) (*string, error) {

	reqParams := new(server.CreateBlockStorageInstanceRequest)
	reqParams.BlockStorageSize = ncloud.Int64(int64(s.Config.BlockStorageSize))
	reqParams.ServerInstanceNo = &serverInstanceNo

	resp, err := s.Conn.server.V2Api.CreateBlockStorageInstance(reqParams)
	if err != nil {
		return nil, err
	}

	blockStorageInstance := resp.BlockStorageInstanceList[0]
	log.Println("Block Storage Instance information : ", blockStorageInstance.BlockStorageInstanceNo)

	if err := waiterBlockStorageInstanceStatus(s.Conn, blockStorageInstance.BlockStorageInstanceNo, "ATTAC", 10*time.Minute); err != nil {
		return nil, errors.New("TIMEOUT : Block Storage instance status is not attached")
	}

	return blockStorageInstance.BlockStorageInstanceNo, nil
}

func (s *StepCreateBlockStorageInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.Config.BlockStorageSize == 0 {
		return processStepResult(nil, s.Error, state)
	}

	s.Say("Create extra block storage instance")

	serverInstanceNo := state.Get("InstanceNo").(string)

	blockStorageInstanceNo, err := s.CreateBlockStorageInstance(serverInstanceNo)
	if err == nil {
		state.Put("BlockStorageInstanceNo", *blockStorageInstanceNo)
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
		reqParams := server.DeleteBlockStorageInstancesRequest{
			BlockStorageInstanceNoList: []*string{blockStorageInstanceNo.(*string)},
		}
		blockStorageInstanceList, err := s.Conn.server.V2Api.DeleteBlockStorageInstances(&reqParams)
		if err != nil {
			return
		}

		s.Say(fmt.Sprintf("Block Storage Instance is deleted. Block Storage InstanceNo is %s", blockStorageInstanceNo.(string)))
		log.Println("Block Storage Instance information : ", blockStorageInstanceList.BlockStorageInstanceList[0])

		if err := waiterBlockStorageInstanceStatus(s.Conn, blockStorageInstanceNo.(*string), "DETAC", time.Minute); err != nil {
			s.Say("TIMEOUT : Block Storage instance status is not deattached")
		}
	}
}
