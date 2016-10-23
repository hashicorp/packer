package googlecompute

import (
	"crypto/rsa"
	"time"
)

// Driver is the interface that has to be implemented to communicate
// with GCE. The Driver interface exists mostly to allow a mock implementation
// to be used to test the steps.
type Driver interface {
	// CreateImage creates an image from the given disk in Google Compute
	// Engine.
	CreateImage(name, description, family, zone, disk string) (<-chan *Image, <-chan error)

	// DeleteImage deletes the image with the given name.
	DeleteImage(name string) <-chan error

	// DeleteInstance deletes the given instance, keeping the boot disk.
	DeleteInstance(zone, name string) (<-chan error, error)

	// DeleteDisk deletes the disk with the given name.
	DeleteDisk(zone, name string) (<-chan error, error)

	// GetImage gets an image; tries the default and public projects.
	GetImage(name string) (*Image, error)

	// GetImageFromProject gets an image from a specific project.
	GetImageFromProject(project, name string) (*Image, error)

	// GetInstanceMetadata gets a metadata variable for the instance, name.
	GetInstanceMetadata(zone, name, key string) (string, error)

	// GetInternalIP gets the GCE-internal IP address for the instance.
	GetInternalIP(zone, name string) (string, error)

	// GetNatIP gets the NAT IP address for the instance.
	GetNatIP(zone, name string) (string, error)

	// GetSerialPortOutput gets the Serial Port contents for the instance.
	GetSerialPortOutput(zone, name string) (string, error)

	// ImageExists returns true if the specified image exists. If an error
	// occurs calling the API, this method returns false.
	ImageExists(name string) bool

	// RunInstance takes the given config and launches an instance.
	RunInstance(*InstanceConfig) (<-chan error, error)

	// WaitForInstance waits for an instance to reach the given state.
	WaitForInstance(state, zone, name string) <-chan error

	// CreateOrResetWindowsPassword creates or resets the password for a user on an Windows instance.
	CreateOrResetWindowsPassword(zone, name string, config *WindowsPasswordConfig) (<-chan error, error)
}

type InstanceConfig struct {
	Address             string
	Description         string
	DiskSizeGb          int64
	DiskType            string
	Image               *Image
	MachineType         string
	Metadata            map[string]string
	Name                string
	Network             string
	OmitExternalIP      bool
	Preemptible         bool
	Region              string
	Scopes              []string
	ServiceAccountEmail string
	Subnetwork          string
	Tags                []string
	Zone                string
}

// WindowsPasswordConfig is the data structue that GCE needs to encrypt the created
// windows password.
type WindowsPasswordConfig struct {
	key      *rsa.PrivateKey
	password string
	UserName string    `json:"userName"`
	Modulus  string    `json:"modulus"`
	Exponent string    `json:"exponent"`
	Email    string    `json:"email"`
	ExpireOn time.Time `json:"expireOn"`
}

type windowsPasswordResponse struct {
	UserName          string `json:"userName"`
	PasswordFound     bool   `json:"passwordFound"`
	EncryptedPassword string `json:"encryptedPassword"`
	Modulus           string `json:"modulus"`
	Exponent          string `json:"exponent"`
	ErrorMessage      string `json:"errorMessage"`
}
