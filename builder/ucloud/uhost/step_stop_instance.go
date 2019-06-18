package uhost

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/common/retry"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

type stepStopInstance struct {
}

func (s *stepStopInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*UCloudClient)
	conn := client.uhostconn
	instance := state.Get("instance").(*uhost.UHostInstanceSet)
	ui := state.Get("ui").(packer.Ui)

	instance, err := client.describeUHostById(instance.UHostId)
	if err != nil {
		return halt(state, err, fmt.Sprintf("Error on reading instance when stop %q", instance.UHostId))
	}

	if instance.State != instanceStateStopped {
		stopReq := conn.NewPoweroffUHostInstanceRequest()
		stopReq.UHostId = ucloud.String(instance.UHostId)
		ui.Say(fmt.Sprintf("Stopping instance %q", instance.UHostId))
		err = retry.Config{
			Tries: 5,
			ShouldRetry: func(err error) bool {
				return err != nil
			},
			RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			if _, err = conn.PoweroffUHostInstance(stopReq); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return halt(state, err, fmt.Sprintf("Error on stopping instance %q", instance.UHostId))
		}

		err = retry.Config{
			Tries: 20,
			ShouldRetry: func(err error) bool {
				return isExpectedStateError(err)
			},
			RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			instance, err := client.describeUHostById(instance.UHostId)
			if err != nil {
				return err
			}

			if instance.State != instanceStateStopped {
				return newExpectedStateError("instance", instance.UHostId)
			}

			return nil
		})

		if err != nil {
			return halt(state, err, fmt.Sprintf("Error on waiting for instance %q to stopped", instance.UHostId))
		}

		ui.Message(fmt.Sprintf("Stop instance %q complete", instance.UHostId))
	}

	return multistep.ActionContinue
}

func (s *stepStopInstance) Cleanup(multistep.StateBag) {
}
