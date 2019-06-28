package uaccount

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UAccountClient is the client of UAccount
type UAccountClient struct {
	*ucloud.Client
}

// NewClient will return a instance of UAccountClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *UAccountClient {
	client := ucloud.NewClient(config, credential)
	return &UAccountClient{
		client,
	}
}
