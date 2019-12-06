/*
Package request is the request of service
*/
package request

import "time"

// Common is the common request
type Common interface {
	GetAction() string
	SetAction(string) error

	GetActionRef() *string
	SetActionRef(*string) error

	GetRegion() string
	SetRegion(string) error

	GetRegionRef() *string
	SetRegionRef(*string) error

	GetProjectId() string
	SetProjectId(string) error

	GetProjectIdRef() *string
	SetProjectIdRef(*string) error

	GetZone() string
	SetZone(string) error

	GetZoneRef() *string
	SetZoneRef(*string) error

	SetRetryCount(int)
	GetRetryCount() int

	WithRetry(int)
	GetMaxretries() int

	WithTimeout(time.Duration)
	GetTimeout() time.Duration

	SetRequestTime(time.Time)
	GetRequestTime() time.Time

	SetRetryable(retryable bool)
	GetRetryable() bool
}

// CommonBase is the base struct of common request
type CommonBase struct {
	Action    *string
	Region    *string
	Zone      *string
	ProjectId *string

	maxRetries  int
	retryable   bool
	retryCount  int
	timeout     time.Duration
	requestTime time.Time
}

// SetRetryCount will set retry count of request
func (c *CommonBase) SetRetryCount(retryCount int) {
	c.retryCount = retryCount
}

// GetRetryCount will return retry count of request
func (c *CommonBase) GetRetryCount() int {
	return c.retryCount
}

// WithRetry will set max retry count of request
func (c *CommonBase) WithRetry(maxRetries int) {
	c.maxRetries = maxRetries
	c.retryable = true
}

// GetMaxretries will return max retry count of request
func (c *CommonBase) GetMaxretries() int {
	return c.maxRetries
}

// SetRetryable will set if the request is retryable
func (c *CommonBase) SetRetryable(retryable bool) {
	c.retryable = retryable
}

// GetRetryable will return if the request is retryable
func (c *CommonBase) GetRetryable() bool {
	return c.retryable
}

// WithTimeout will set timeout of request
func (c *CommonBase) WithTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout will get timeout of request
func (c *CommonBase) GetTimeout() time.Duration {
	return c.timeout
}

// SetRequestTime will set timeout of request
func (c *CommonBase) SetRequestTime(requestTime time.Time) {
	c.requestTime = requestTime
}

// GetRequestTime will get timeout of request
func (c *CommonBase) GetRequestTime() time.Time {
	return c.requestTime
}

// GetAction will return action of request
func (c *CommonBase) GetAction() string {
	if c.Action == nil {
		return ""
	}
	return *c.Action
}

// SetAction will set action of request
func (c *CommonBase) SetAction(val string) error {
	c.Action = &val
	return nil
}

// GetActionRef will return a pointer to action of request
func (c *CommonBase) GetActionRef() *string {
	return c.Action
}

// SetActionRef will set a pointer to action of request
func (c *CommonBase) SetActionRef(val *string) error {
	c.Action = val
	return nil
}

// GetRegion will return region of request
func (c *CommonBase) GetRegion() string {
	if c.Region == nil {
		return ""
	}
	return *c.Region
}

// SetRegion will set region of request
func (c *CommonBase) SetRegion(val string) error {
	c.Region = &val
	return nil
}

// GetRegionRef will return a pointer to region of request
func (c *CommonBase) GetRegionRef() *string {
	return c.Region
}

// SetRegionRef will set a pointer to region of request
func (c *CommonBase) SetRegionRef(val *string) error {
	c.Region = val
	return nil
}

// GetZone will return zone of request
func (c *CommonBase) GetZone() string {
	if c.Zone == nil {
		return ""
	}
	return *c.Zone
}

// SetZone will set zone of request
func (c *CommonBase) SetZone(val string) error {
	c.Zone = &val
	return nil
}

// GetZoneRef will return a pointer to zone of request
func (c *CommonBase) GetZoneRef() *string {
	return c.Zone
}

// SetZoneRef will set a pointer to zone of request
func (c *CommonBase) SetZoneRef(val *string) error {
	c.Zone = val
	return nil
}

// GetProjectId will get project id of request
func (c *CommonBase) GetProjectId() string {
	if c.ProjectId == nil {
		return ""
	}
	return *c.ProjectId
}

// SetProjectId will set project id of request
func (c *CommonBase) SetProjectId(val string) error {
	c.ProjectId = &val
	return nil
}

// GetProjectIdRef will get a pointer to project id of request
func (c *CommonBase) GetProjectIdRef() *string {
	return c.ProjectId
}

// SetProjectIdRef will set a pointer to project id of request
func (c *CommonBase) SetProjectIdRef(val *string) error {
	c.ProjectId = val
	return nil
}
