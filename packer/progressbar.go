package packer

import (
	"io"

	"github.com/cheggaaa/pb"
)

// ProgressBar allows to graphically display
// a self refreshing progress bar.
// No-op When in machine readable mode.
type ProgressBar interface {
	Start(total uint64)
	Add(current uint64)
	NewProxyReader(r io.Reader) (proxy io.Reader)
	Finish()
}

type BasicProgressBar struct {
	*pb.ProgressBar
}

func (bpb *BasicProgressBar) Start(total uint64) {
	bpb.SetTotal64(int64(total))
	bpb.ProgressBar.Start()
}

func (bpb *BasicProgressBar) Add(current uint64) {
	bpb.ProgressBar.Add64(int64(current))
}
func (bpb *BasicProgressBar) NewProxyReader(r io.Reader) io.Reader {
	return &ProxyReader{
		Reader:      r,
		ProgressBar: bpb,
	}
}
func (bpb *BasicProgressBar) NewProxyReadCloser(r io.ReadCloser) io.ReadCloser {
	return &ProxyReader{
		Reader:      r,
		ProgressBar: bpb,
	}
}

var _ ProgressBar = new(BasicProgressBar)

// NoopProgressBar is a silent progress bar
type NoopProgressBar struct {
}

func (npb *NoopProgressBar) Start(uint64)                                     {}
func (npb *NoopProgressBar) Add(uint64)                                       {}
func (npb *NoopProgressBar) Finish()                                          {}
func (npb *NoopProgressBar) NewProxyReader(r io.Reader) io.Reader             { return r }
func (npb *NoopProgressBar) NewProxyReadCloser(r io.ReadCloser) io.ReadCloser { return r }

var _ ProgressBar = new(NoopProgressBar)

// ProxyReader implements io.ReadCloser but sends
// count of read bytes to progress bar
type ProxyReader struct {
	io.Reader
	ProgressBar
}

func (r *ProxyReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.ProgressBar.Add(uint64(n))
	return
}

// Close the reader if it implements io.Closer
func (r *ProxyReader) Close() (err error) {
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return
}
