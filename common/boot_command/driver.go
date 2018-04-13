package bootcommand

const shiftedChars = "~!@#$%^&*()_+{}|:\"<>?"

// BCDriver is our access to the VM we want to type boot commands to
type BCDriver interface {
	SendKey(key rune, action KeyAction) error
	SendSpecial(special string, action KeyAction) error
	// Finalize will be called after every expression has been processed.
	Finalize() error
}
