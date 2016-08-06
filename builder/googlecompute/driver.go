package googlecompute

// Driver is the interface that has to be implemented to communicate
// with GCE. The Driver interface exists mostly to allow a mock implementation
// to be used to test the steps.
type Driver interface {
	// ImageExists returns true if the specified image exists. If an error
	// occurs calling the API, this method returns false.
	ImageExists(name string) bool

	// CreateImage creates an image from the given disk in Google Compute
	// Engine.
	CreateImage(name, description, family, zone, disk string) (<-chan Image, <-chan error)

	// DeleteImage deletes the image with the given name.
	DeleteImage(name string) <-chan error

	// DeleteInstance deletes the given instance, keeping the boot disk.
	DeleteInstance(zone, name string) (<-chan error, error)

	// DeleteDisk deletes the disk with the given name.
	DeleteDisk(zone, name string) (<-chan error, error)

	// GetInternalIP gets the GCE-internal IP address for the instance.
	GetInternalIP(zone, name string) (string, error)

	// GetNatIP gets the NAT IP address for the instance.
	GetNatIP(zone, name string) (string, error)
	
	// GetSerialPortOutput gets the Serial Port contents for the instance.
	GetSerialPortOutput(zone, name string) (string, error)

	// RunInstance takes the given config and launches an instance.
	RunInstance(*InstanceConfig) (<-chan error, error)

	// WaitForInstance waits for an instance to reach the given state.
	WaitForInstance(state, zone, name string) <-chan error
}

type Image struct {
	Name      string
	ProjectId string
	SizeGb    int64
}

type InstanceConfig struct {
	Address             string
	Description         string
	DiskSizeGb          int64
	DiskType            string
	Image               Image
	MachineType         string
	Metadata            map[string]string
	Name                string
	Network             string
	OmitExternalIP      bool
	Preemptible         bool
	Region              string
	ServiceAccountEmail string
	Subnetwork          string
	Tags                []string
	Zone                string
}
