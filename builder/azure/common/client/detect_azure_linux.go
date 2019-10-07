package client

import (
	"bytes"
	"io/ioutil"
)

var (
	smbiosAssetTagFile = "/sys/class/dmi/id/chassis_asset_tag"
	azureAssetTag      = []byte("7783-7084-3265-9085-8269-3286-77\n")
)

// IsAzure returns true if Packer is running on Azure
func IsAzure() bool {
	return isAzureAssetTag(smbiosAssetTagFile)
}

func isAzureAssetTag(filename string) bool {
	if d, err := ioutil.ReadFile(filename); err == nil {
		return bytes.Compare(d, azureAssetTag) == 0
	}
	return false
}
