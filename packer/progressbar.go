package packer

import (
	"fmt"
	"io"
	"sync"

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
	mtx   sync.Mutex // locks in Start, Finish, Add & NewProxyReader
	Bar   BasicProgressBar
	items int32
	total int64

	started             bool
	ConfigProgressbarFN func(*pb.ProgressBar)
}

var _ ProgressBar = new(StackableProgressBar)

func defaultProgressbarConfigFn(bar *pb.ProgressBar) {
	bar.SetUnits(pb.U_BYTES)
}

func (spb *StackableProgressBar) start() {
	bar := pb.New(0)
	if spb.ConfigProgressbarFN == nil {
		spb.ConfigProgressbarFN = defaultProgressbarConfigFn
	}
	spb.ConfigProgressbarFN(bar)

	bar.Start()
	spb.Bar.ProgressBar = bar
	spb.started = true
}

func (spb *StackableProgressBar) Start(total int64) {
	spb.mtx.Lock()

	spb.total += total
	spb.items++

	if !spb.started {
		spb.start()
	}
	spb.Bar.SetTotal64(spb.total)
	spb.prefix()
	spb.mtx.Unlock()
}

func (spb *StackableProgressBar) Add(total int64) {
	spb.mtx.Lock()
	defer spb.mtx.Unlock()
	if spb.Bar.ProgressBar != nil {
		spb.Bar.Add(total)
	}
}

func (spb *StackableProgressBar) NewProxyReader(r io.Reader) io.Reader {
	spb.mtx.Lock()
	defer spb.mtx.Unlock()
	return spb.Bar.NewProxyReader(r)
}

func (spb *StackableProgressBar) prefix() {
	spb.Bar.ProgressBar.Prefix(fmt.Sprintf("%d items: ", spb.items))
}

func (spb *StackableProgressBar) Finish() {
	spb.mtx.Lock()
	defer spb.mtx.Unlock()

	if spb.items < 0 {
		spb.items--
	}
	if spb.items == 0 && spb.Bar.ProgressBar != nil {
		// slef cleanup
		spb.Bar.ProgressBar.Finish()
		spb.Bar.ProgressBar = nil
		spb.started = false
		spb.total = 0
		return
	}
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
