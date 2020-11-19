package packer

import "github.com/hashicorp/go-getter"

// The Ui interface handles all communication for Packer with the outside
// world. This sort of control allows us to strictly control how output
// is formatted and various levels of output.
type Ui interface {
	Ask(string) (string, error)
	Say(string)
	Message(string)
	Error(string)
	Machine(string, ...string)
	// TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser)
	getter.ProgressTracker
}
