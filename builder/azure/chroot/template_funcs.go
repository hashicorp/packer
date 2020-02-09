package chroot

import (
	"fmt"
	"sync"

	"github.com/hashicorp/packer/builder/azure/common/client"
)

// CreateVMMetadataTemplateFunc returns a template function that retrieves VM metadata. VM metadata is retrieved only once and reused for all executions of the function.
func CreateVMMetadataTemplateFunc() func(string) (string, error) {
	var data *client.ComputeInfo
	var dataErr error
	once := sync.Once{}
	return func(key string) (string, error) {
		once.Do(func() {
			data, dataErr = client.DefaultMetadataClient.GetComputeInfo()
		})
		if dataErr != nil {
			return "", dataErr
		}
		switch key {
		case "name":
			return data.Name, nil
		case "subscription_id":
			return data.SubscriptionID, nil
		case "resource_group":
			return data.ResourceGroupName, nil
		case "location":
			return data.Location, nil
		case "resource_id":
			return data.ResourceID(), nil
		default:
			return "", fmt.Errorf("unknown metadata key: %s (supported: name, subscription_id, resource_group, location, resource_id)", key)
		}
	}
}
