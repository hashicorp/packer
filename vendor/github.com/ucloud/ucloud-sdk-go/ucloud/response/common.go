/*
Package response is the response of service
*/
package response

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

// Common describe a response of action,
// it is only used for ucloud open api v1 via HTTP GET and action parameters.
type Common interface {
	GetRetCode() int
	GetMessage() string
	GetAction() string

	GetRequest() request.Common
	SetRequest(request.Common)

	SetRequestUUID(string)
	GetRequestUUID() string
}

// CommonBase has common attribute and method,
// it also implement ActionResponse interface.
type CommonBase struct {
	Action  string
	RetCode int
	Message string

	requestUUID string

	request request.Common
}

// GetRetCode will return the error code of ucloud api
// Error is non-zero and success is zero
func (c *CommonBase) GetRetCode() int {
	return c.RetCode
}

// GetMessage will return the error message of ucloud api
func (c *CommonBase) GetMessage() string {
	return c.Message
}

// GetAction will return the request action of ucloud api
func (c *CommonBase) GetAction() string {
	return c.Action
}

// GetRequest will return the latest retried request of current action
func (c *CommonBase) GetRequest() request.Common {
	return c.request
}

// GetRequest will return the latest retried request of current action
func (c *CommonBase) SetRequest(req request.Common) {
	c.request = req
}

// SetRequestUUID will set uuid of request
func (c *CommonBase) SetRequestUUID(uuid string) {
	c.requestUUID = uuid
}

// GetRequestUUID will get uuid of request
func (c *CommonBase) GetRequestUUID() string {
	return c.requestUUID
}
