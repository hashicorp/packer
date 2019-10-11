package uhost

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UHostClient is the client of UHost
type UHostClient struct {
	*ucloud.Client
}

// NewClient will return a instance of UHostClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *UHostClient {
	client := ucloud.NewClient(config, credential)
	return &UHostClient{
		client,
	}
}
