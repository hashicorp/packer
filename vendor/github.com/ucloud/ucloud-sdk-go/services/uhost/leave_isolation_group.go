//Code is generated by ucloud code generator, don't modify it by hand, it will cause undefined behaviors.
//go:generate ucloud-gen-go-api UHost LeaveIsolationGroup

package uhost

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// LeaveIsolationGroupRequest is request schema for LeaveIsolationGroup action
type LeaveIsolationGroupRequest struct {
	request.CommonBase

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// [公共参数] 可用区信息
	// Zone *string `required:"false"`

	// [公共参数] 项目id
	// ProjectId *string `required:"false"`

	// 硬件隔离组id
	GroupId *string `required:"true"`

	// 主机id
	UHostId *string `required:"true"`
}

// LeaveIsolationGroupResponse is response schema for LeaveIsolationGroup action
type LeaveIsolationGroupResponse struct {
	response.CommonBase

	// 主机id
	UHostId string
}

// NewLeaveIsolationGroupRequest will create request of LeaveIsolationGroup action.
func (c *UHostClient) NewLeaveIsolationGroupRequest() *LeaveIsolationGroupRequest {
	req := &LeaveIsolationGroupRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

// LeaveIsolationGroup - 移除硬件隔离组中的主机
func (c *UHostClient) LeaveIsolationGroup(req *LeaveIsolationGroupRequest) (*LeaveIsolationGroupResponse, error) {
	var err error
	var res LeaveIsolationGroupResponse

	err = c.Client.InvokeAction("LeaveIsolationGroup", req, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
