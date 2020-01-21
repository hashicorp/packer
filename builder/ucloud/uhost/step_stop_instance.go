package uhost

import (
	"context"
	"fmt"
	"time"

	ucloudcommon "github.com/hashicorp/packer/builder/ucloud/common"
	"github.com/hashicorp/packer/common/retry"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

type stepStopInstance struct {
}

func (s *stepStopInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ucloudcommon.UCloudClient)
	conn := client.UHostConn
	instance := state.Get("instance").(*uhost.UHostInstanceSet)
	ui := state.Get("ui").(packer.Ui)

	instance, err := client.DescribeUHostById(instance.UHostId)
	if err != nil {
		return ucloudcommon.Halt(state, err, fmt.Sprintf("Error on reading instance when stopping %q", instance.UHostId))
	}

	if instance.State != ucloudcommon.InstanceStateStopped {
		stopReq := conn.NewStopUHostInstanceRequest()
		stopReq.UHostId = ucloud.String(instance.UHostId)
		ui.Say(fmt.Sprintf("Stopping instance %q", instance.UHostId))
		err = retry.Config{
			Tries: 5,
			ShouldRetry: func(err error) bool {
				return err != nil
			},
			RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			if _, err = conn.StopUHostInstance(stopReq); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return ucloudcommon.Halt(state, err, fmt.Sprintf("Error on stopping instance %q", instance.UHostId))
		}

		err = retry.Config{
			Tries: 20,
			ShouldRetry: func(err error) bool {
				return ucloudcommon.IsExpectedStateError(err)
			},
			RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			instance, err := client.DescribeUHostById(instance.UHostId)
			if err != nil {
				return err
			}

			if instance.State != ucloudcommon.InstanceStateStopped {
				return ucloudcommon.NewExpectedStateError("instance", instance.UHostId)
			}

			return nil
		})

		if err != nil {
			return ucloudcommon.Halt(state, err, fmt.Sprintf("Error on waiting for stopping instance when stopping %q", instance.UHostId))
		}

		ui.Message(fmt.Sprintf("Stopping instance %q complete", instance.UHostId))
	}

	return multistep.ActionContinue
}

func (s *stepStopInstance) Cleanup(multistep.StateBag) {
}
