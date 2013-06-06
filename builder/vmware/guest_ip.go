package vmware

// Interface to help find the IP address of a running virtual machine.
type GuestIPFinder interface {
	GuestIP() (string, error)
}

// DHCPLeaseGuestLookup looks up the IP address of a guest using DHCP
// lease information from the VMware network devices.
type DHCPLeaseGuestLookup struct {
	// Device that the guest is connected to.
	Device     string

	// MAC address of the guest.
	MACAddress string
}

func (f *DHCPLeaseGuestLookup) GuestIP() (string, error) {
	return "", nil
}
