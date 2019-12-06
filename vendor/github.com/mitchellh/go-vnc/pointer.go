package vnc

// ButtonMask represents a mask of pointer presses/releases.
type ButtonMask uint8

// All available button mask components.
const (
	ButtonLeft ButtonMask = 1 << iota
	ButtonMiddle
	ButtonRight
	Button4
	Button5
	Button6
	Button7
	Button8
)
