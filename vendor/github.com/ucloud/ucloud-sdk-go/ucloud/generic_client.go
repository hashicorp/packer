package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

func (c *Client) NewGenericRequest() request.GenericRequest {
	req := &request.BaseGenericRequest{}

	// setup request with client config
	c.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(false)
	return req
}

func (c *Client) GenericInvoke(req request.GenericRequest) (response.GenericResponse, error) {
	var res response.BaseGenericResponse

	err := c.InvokeAction(req.GetAction(), req, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
