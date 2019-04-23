package googlecompute

import compute "google.golang.org/api/compute/v1"

// Configis the interface that has to be implemented to pass
//configuration for GCE. The Config interface exists mostly to allow a mock implementation
// to be used to test the steps.
type Config interface {
	//GetImageDescription gets image desription for the image.
	GetImageDescription() string

	//GetImageName gets image name for the image.
	GetImageName() string

	//GGetImageFamily gets image family for the image.
	GetImageFamily() string

	//GetImageLabels gets list of labels for the image.
	GetImageLabels() map[string]string

	//GetImageDescription gets image desription for the image.
	GetImageLicenses() []string

	//GetZone gets image zone for the image.
	GetZone() string

	//GetDiskName gets disk name for the image.
	GetDiskName() string

	//GetImageEncryptionKey gets image encryption key for the image.
	GetImageEncryptionKey() *compute.CustomerEncryptionKey
}
