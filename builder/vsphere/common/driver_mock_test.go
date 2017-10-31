package common

import (
	"testing"
)

func TestDriverMock_impl(t *testing.T) {
	var _ Driver = new(DriverMock)
}
