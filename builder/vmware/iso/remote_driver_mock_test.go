package iso

import (
	"testing"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
)

func TestRemoteDriverMock_impl(t *testing.T) {
	var _ vmwcommon.Driver = new(RemoteDriverMock)
	var _ RemoteDriver = new(RemoteDriverMock)
}
