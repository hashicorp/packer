package communicator

// WinRMConfig is configuration that can be returned at runtime to
// dynamically configure WinRM.
type WinRMConfig struct {
	Username string
	Password string
}
