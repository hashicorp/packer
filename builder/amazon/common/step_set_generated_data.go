package common

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/common/packerbuilderdata"
	"github.com/hashicorp/packer/helper/multistep"
)

// &awscommon.StepSetGeneratedData{
// 	GeneratedData: generatedData,
// },

type StepSetGeneratedData struct {
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepSetGeneratedData) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)

	extractBuildInfo(*ec2conn.Config.Region, state, s.GeneratedData)

	return multistep.ActionContinue
}

func (s *StepSetGeneratedData) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
