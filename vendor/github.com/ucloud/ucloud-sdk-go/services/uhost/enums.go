package uhost

// State is the state of UHost instance
type State string

// Enum values for State
const (
	StateInitializing State = "Initializing"
	StateStarting     State = "Starting"
	StateRunning      State = "Running"
	StateStopping     State = "Stopping"
	StateStopped      State = "Stopped"
	StateInstallFail  State = "InstallFail"
	StateRebooting    State = "Rebooting"
)

// MarshalValue will marshal state value to string
func (enum State) MarshalValue() (string, error) {
	return string(enum), nil
}
