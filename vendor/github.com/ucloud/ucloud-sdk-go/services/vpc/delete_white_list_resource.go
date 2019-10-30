// Code is generated by ucloud-model, DO NOT EDIT IT.

package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// DeleteWhiteListResourceRequest is request schema for DeleteWhiteListResource action
type DeleteWhiteListResourceRequest struct {
	request.CommonBase

	// [公共参数] 项目Id。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	// ProjectId *string `required:"false"`

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// NAT网关Id
	NATGWId *string `required:"true"`

	// 删除白名单的资源Id
	ResourceIds []string `required:"true"`
}

// DeleteWhiteListResourceResponse is response schema for DeleteWhiteListResource action
type DeleteWhiteListResourceResponse struct {
	response.CommonBase
}

// NewDeleteWhiteListResourceRequest will create request of DeleteWhiteListResource action.
func (c *VPCClient) NewDeleteWhiteListResourceRequest() *DeleteWhiteListResourceRequest {
	req := &DeleteWhiteListResourceRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

// DeleteWhiteListResource - 删除NAT网关白名单列表
func (c *VPCClient) DeleteWhiteListResource(req *DeleteWhiteListResourceRequest) (*DeleteWhiteListResourceResponse, error) {
	var err error
	var res DeleteWhiteListResourceResponse

	reqCopier := *req

	err = c.Client.InvokeAction("DeleteWhiteListResource", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
