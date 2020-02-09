package qemu

import (
	"context"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
)

// Step to discover the http ip
// which guests use to reach the vm host
// To make sure the IP is set before boot command and http server steps
type stepHTTPIPDiscover struct{}

func (s *stepHTTPIPDiscover) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	hostIP := "10.0.2.2"
	common.SetHTTPIP(hostIP)

	return multistep.ActionContinue
}

func (s *stepHTTPIPDiscover) Cleanup(state multistep.StateBag) {}
