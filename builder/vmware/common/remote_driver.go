package common

import (
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type RemoteDriver interface {
	Driver

	// UploadISO uploads a local ISO to the remote side and returns the
	// new path that should be used in the VMX along with an error if it
	// exists.
	UploadISO(path string, checksum string, ui packersdk.Ui) (string, error)

	// RemoveCache deletes localPath from the remote cache.
	RemoveCache(localPath string) error

	// Adds a VM to inventory specified by the path to the VMX given.
	Register(path string) error

	// Removes a VM from inventory specified by the path to the VMX given.
	Unregister(path string) error

	// Destroys a VM
	Destroy() error

	// Checks if the VM is destroyed.
	IsDestroyed() (bool, error)

	// Uploads a local file to remote side.
	upload(dst, src string, ui packersdk.Ui) error

	// Download a remote file to a local file.
	Download(src, dst string) error

	// Reload VM on remote side.
	ReloadVM() error
}
