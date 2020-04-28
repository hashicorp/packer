package chroot

import "github.com/hashicorp/packer/builder/azure/common/client"

// Diskset represents all of the disks or snapshots associated with an image.
// It maps lun to resource ids. The OS disk is stored with lun=-1.
type Diskset map[int32]client.Resource

// OS return the OS disk resource ID or nil if it is not assigned
func (ds Diskset) OS() *client.Resource {
	if r, ok := ds[-1]; ok {
		return &r
	}
	return nil
}

// Data return the data disk resource ID or nil if it is not assigned
func (ds Diskset) Data(lun int32) *client.Resource {
	if r, ok := ds[lun]; ok {
		return &r
	}
	return nil
}
