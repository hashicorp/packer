package common

import (
	"testing"
)

func TestRemoteDriverMock_impl(t *testing.T) {
	var _ Driver = new(RemoteDriverMock)
	var _ RemoteDriver = new(RemoteDriverMock)
}
