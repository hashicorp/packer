package common

// These are the different valid mode values for "guest_additions_mode" which
// determine how guest additions are delivered to the guest.
const (
	GuestAdditionsModeDisable string = "disable"
	GuestAdditionsModeAttach         = "attach"
	GuestAdditionsModeUpload         = "upload"
)
