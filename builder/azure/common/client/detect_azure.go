// +build !linux

package client

// IsAzure returns true if Packer is running on Azure (currently only works on Linux)
func IsAzure() bool {
	return false
}
