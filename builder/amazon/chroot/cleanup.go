package chroot

// Cleanup is an interface that some steps implement for early cleanup.
type Cleanup interface {
	CleanupFunc(map[string]interface{}) error
}
