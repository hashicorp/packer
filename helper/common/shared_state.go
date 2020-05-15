package common

// This is used in the BasicPlaceholderData() func in the packer/provisioner.go
// To force users to access generated data via the "generated" func.
const PlaceholderMsg = "To set this dynamically in the Packer template, " +
	"you must use the `build` function"
