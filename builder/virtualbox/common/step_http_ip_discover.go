package common

import (
	"context"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
)

// Step to discover the http ip
// which guests use to reach the vm host
// To make sure the IP is set before boot command and http server steps
type StepHTTPIPDiscover struct{}

func (s *StepHTTPIPDiscover) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	hostIP := "10.0.2.2"
	common.SetHTTPIP(hostIP)

	return multistep.ActionContinue
}

func (s *StepHTTPIPDiscover) Cleanup(state multistep.StateBag) {}
