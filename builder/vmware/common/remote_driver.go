package common

type RemoteDriver interface {
	Driver

	// UploadISO uploads a local ISO to the remote side and returns the
	// new path that should be used in the VMX along with an error if it
	// exists.
	UploadISO(string, string, string) (string, error)

	// Adds a VM to inventory specified by the path to the VMX given.
	Register(string) error

	// Removes a VM from inventory specified by the path to the VMX given.
	Unregister(string) error

	// Destroys a VM
	Destroy() error

	// Checks if the VM is destroyed.
	IsDestroyed() (bool, error)

	// Uploads a local file to remote side.
	Upload(dst, src string) error

	// Reload VM on remote side.
	ReloadVM() error

	// Read bytes from of a remote file.
	ReadFile(string) ([]byte, error)

	// Write bytes to a remote file.
	WriteFile(string, []byte) error
}
