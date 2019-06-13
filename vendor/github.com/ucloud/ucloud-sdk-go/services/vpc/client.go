package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// VPCClient is the client of VPC2.0
type VPCClient struct {
	*ucloud.Client
}

// NewClient will return a instance of VPCClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *VPCClient {
	client := ucloud.NewClient(config, credential)
	return &VPCClient{
		client,
	}
}
