package packer

import "io"

// NoopProgressTracker is a progress tracker
// that displays nothing.
type NoopProgressTracker struct{}

// TrackProgress returns stream
func (*NoopProgressTracker) TrackProgress(_ string, _, _ int64, stream io.ReadCloser) io.ReadCloser {
	return stream
}
