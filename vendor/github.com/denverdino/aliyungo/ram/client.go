package ram

import (
	"github.com/denverdino/aliyungo/common"
	"os"
)

const (
	// RAMDefaultEndpoint is the default API endpoint of RAM services
	RAMDefaultEndpoint = "https://ram.aliyuncs.com"
	RAMAPIVersion      = "2015-05-01"
)

type RamClient struct {
	common.Client
}

func NewClient(accessKeyId string, accessKeySecret string) RamClientInterface {
	endpoint := os.Getenv("RAM_ENDPOINT")
	if endpoint == "" {
		endpoint = RAMDefaultEndpoint
	}
	return NewClientWithEndpoint(endpoint, accessKeyId, accessKeySecret)
}

func NewClientWithEndpoint(endpoint string, accessKeyId string, accessKeySecret string) RamClientInterface {
	client := &RamClient{}
	client.Init(endpoint, RAMAPIVersion, accessKeyId, accessKeySecret)
	return client
}
