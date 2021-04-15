package common

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// Step to discover the http ip
// which guests use to reach the vm host
// To make sure the IP is set before boot command and http server steps
type StepHTTPIPDiscover struct{}

func (s *StepHTTPIPDiscover) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	state.Put("http_ip", "10.0.2.2")

	return multistep.ActionContinue
}

func (s *StepHTTPIPDiscover) Cleanup(state multistep.StateBag) {}
