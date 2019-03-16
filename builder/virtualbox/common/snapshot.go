package common

// VBoxSnapshot stores the hierarchy of snapshots for a VM instance
type VBoxSnapshot struct {
	Name      string
	UUID      string
	IsCurrent bool
	Parent    *VBoxSnapshot // nil if topmost (root) snapshot
	Children  []VBoxSnapshot
}

// IsChildOf verifies if the current snaphot is a child of the passed as argument
func (sn *VBoxSnapshot) IsChildOf(candidate *VBoxSnapshot) bool {
	return false
}
