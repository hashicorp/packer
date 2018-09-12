package packer

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/cheggaaa/pb"
)

// ProgressBar allows to graphically display
// a self refreshing progress bar.
type ProgressBar interface {
	Start(total int64)
	Add(current int64)
	NewProxyReader(r io.Reader) (proxy io.Reader)
	Finish()
}

// StackableProgressBar is a progress bar that
// allows to track multiple downloads at once.
// Every call to Start increments a counter that
// will display the number of current loadings.
// Every call to Start will add total to an internal
// total that is the total displayed.
// First call to Start will start a goroutine
// that is waiting for every download to be finished.
// Last call to Finish triggers a cleanup.
// When all active downloads are finished
// StackableProgressBar will clean itself to a default
// state.
type StackableProgressBar struct {
	mtx sync.Mutex // locks in Start & Finish
	BasicProgressBar
	items int32
	total int64

	started bool
}

var _ ProgressBar = new(StackableProgressBar)

func (spb *StackableProgressBar) start() {
	spb.BasicProgressBar.ProgressBar = pb.New(0)
	spb.BasicProgressBar.ProgressBar.SetUnits(pb.U_BYTES)

	spb.BasicProgressBar.ProgressBar.Start()
	spb.started = true
}

func (spb *StackableProgressBar) Start(total int64) {
	spb.mtx.Lock()

	spb.total += total
	spb.items++

	if !spb.started {
		spb.start()
	}
	spb.SetTotal64(spb.total)
	spb.prefix()
	spb.mtx.Unlock()
}

func (spb *StackableProgressBar) prefix() {
	spb.BasicProgressBar.ProgressBar.Prefix(fmt.Sprintf("%d items: ", atomic.LoadInt32(&spb.items)))
}

func (spb *StackableProgressBar) Finish() {
	spb.mtx.Lock()
	defer spb.mtx.Unlock()

	spb.items--
	if spb.items == 0 {
		// slef cleanup
		spb.BasicProgressBar.ProgressBar.Finish()
		spb.BasicProgressBar.ProgressBar = nil
		spb.started = false
		spb.total = 0
		return
	}
	spb.prefix()
}

// BasicProgressBar is packer's basic progress bar.
// Current implementation will always try to keep
// itself at the bottom of a terminal.
type BasicProgressBar struct {
	*pb.ProgressBar
}

var _ ProgressBar = new(BasicProgressBar)

func (bpb *BasicProgressBar) Start(total int64) {
	bpb.SetTotal64(total)
	bpb.ProgressBar.Start()
}

func (bpb *BasicProgressBar) Add(current int64) {
	bpb.ProgressBar.Add64(current)
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

// NoopProgressBar is a silent progress bar.
type NoopProgressBar struct {
}

var _ ProgressBar = new(NoopProgressBar)

func (npb *NoopProgressBar) Start(int64)                                      {}
func (npb *NoopProgressBar) Add(int64)                                        {}
func (npb *NoopProgressBar) Finish()                                          {}
func (npb *NoopProgressBar) NewProxyReader(r io.Reader) io.Reader             { return r }
func (npb *NoopProgressBar) NewProxyReadCloser(r io.ReadCloser) io.ReadCloser { return r }

// ProxyReader implements io.ReadCloser but sends
// count of read bytes to a progress bar
type ProxyReader struct {
	io.Reader
	ProgressBar
}

func (r *ProxyReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.ProgressBar.Add(int64(n))
	return
}

// Close the reader if it implements io.Closer
func (r *ProxyReader) Close() (err error) {
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return
}
