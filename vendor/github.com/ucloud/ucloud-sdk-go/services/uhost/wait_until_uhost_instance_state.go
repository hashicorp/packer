package uhost

import (
	"time"

	"github.com/ucloud/ucloud-sdk-go/private/utils"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

// WaitUntilUHostInstanceStateRequest is the request of uhost instance state waiter
type WaitUntilUHostInstanceStateRequest struct {
	request.CommonBase

	Interval        *time.Duration
	MaxAttempts     *int
	DescribeRequest *DescribeUHostInstanceRequest
	State           State
	IgnoreError     *bool
}

// NewWaitUntilUHostInstanceStateRequest will create request of WaitUntilUHostInstanceState action.
func (c *UHostClient) NewWaitUntilUHostInstanceStateRequest() *WaitUntilUHostInstanceStateRequest {
	cfg := c.Client.GetConfig()

	return &WaitUntilUHostInstanceStateRequest{
		CommonBase: request.CommonBase{
			Region:    ucloud.String(cfg.Region),
			ProjectId: ucloud.String(cfg.ProjectId),
		},
	}
}

// WaitUntilUHostInstanceState will pending current goroutine until the state has changed to expected state.
func (c *UHostClient) WaitUntilUHostInstanceState(req *WaitUntilUHostInstanceStateRequest) error {
	waiter := utils.FuncWaiter{
		Interval:    ucloud.TimeDurationValue(req.Interval),
		MaxAttempts: ucloud.IntValue(req.MaxAttempts),
		IgnoreError: ucloud.BoolValue(req.IgnoreError),
		Checker: func() (bool, error) {
			resp, err := c.DescribeUHostInstance(req.DescribeRequest)

			if err != nil {
				skipErrors := []string{uerr.ErrNetwork, uerr.ErrHTTPStatus, uerr.ErrRetCode}
				if uErr, ok := err.(uerr.Error); ok && utils.IsStringIn(uErr.Name(), skipErrors) {
					log.Infof("skip error for wait resource state, %s", uErr)
					return false, nil
				}
				log.Infof("wait for resource state is ready, %s", err)
				return false, err
			}

			// TODO: Ensure if it is any data consistency problem?
			// Such as creating a new uhost, but cannot describe it's correct state immediately ...
			for _, uhost := range resp.UHostSet {
				if val, _ := req.State.MarshalValue(); uhost.State != val {
					return false, nil
				}
			}

			if len(resp.UHostSet) > 0 {
				return true, nil
			}

			return false, nil
		},
	}
	return waiter.WaitForCompletion()
}
