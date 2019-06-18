package unet

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UNetClient is the client of UNet
type UNetClient struct {
	*ucloud.Client
}

// NewClient will return a instance of UNetClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *UNetClient {
	client := ucloud.NewClient(config, credential)
	return &UNetClient{
		client,
	}
}
