package command

import (
	"context"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
)

type Scheduler interface {
	Build(*BuildArgs) int
	Validate(*ValidateArgs) int
	Inspect(*InspectArgs) int
}

// NewScheduler returns a new scheduler for running commands with.
func NewScheduler(
	cfg packer.Handler,
	ui packersdk.Ui,
	context context.Context,
) Scheduler {
	return NewSequentialScheduler(cfg, ui, context)
}
