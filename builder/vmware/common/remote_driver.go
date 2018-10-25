package common

type RemoteDriver interface {
	Driver

	// UploadISO uploads a local ISO to the remote side and returns the
	// new path that should be used in the VMX along with an error if it
	// exists.
	UploadISO(string, string, string) (string, error)

	// RemoveCache deletes localPath from the remote cache.
	RemoveCache(localPath string) error

	// Adds a VM to inventory specified by the path to the VMX given.
	Register(string) error

	// Removes a VM from inventory specified by the path to the VMX given.
	Unregister(string) error

	// Destroys a VM
	Destroy() error

	// Checks if the VM is destroyed.
	IsDestroyed() (bool, error)

	// Uploads a local file to remote side.
	upload(dst, src string) error

	// Download a remote file to a local file.
	Download(src, dst string) error

	// Reload VM on remote side.
	ReloadVM() error
}
