package bootcommand

const shiftedChars = "~!@#$%^&*()_+{}|:\"<>?"

// RebootFunc will be called to reboot the VM
type RebootFunc func() error

func noopReboot() error {
	return nil
}

// BCDriver is our access to the VM we want to type boot commands to
type BCDriver interface {
	Reboot() error
	SendKey(key rune, action KeyAction) error
	SendSpecial(special string, action KeyAction) error
	// Flush will be called when we want to send scancodes to the VM.
	Flush() error
}
