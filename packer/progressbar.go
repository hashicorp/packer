package packer

import (
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

// StackableProgressBar is a progress bar pool that
// allows to track multiple advencments at once,
// by displaying multiple bars.
type StackableProgressBar struct {
	mtx sync.Mutex // locks in Start & Finish

	pool *pb.Pool
	bars []*BasicProgressBar

	wg sync.WaitGroup
}

func (spb *StackableProgressBar) cleanup() {
	spb.wg.Wait()

	spb.mtx.Lock()
	defer spb.mtx.Unlock()

	spb.pool.Stop()
	spb.pool = nil
	spb.bars = nil
}

func (spb *StackableProgressBar) New(identifier string) ProgressBar {
	spb.mtx.Lock()
	spb.wg.Add(1)
	defer spb.mtx.Unlock()

	start := false
	if spb.pool == nil {
		spb.pool = pb.NewPool()
		go spb.cleanup()
		start = true
	}

	bar := NewProgressBar(identifier)
	bar.Prefix(identifier)
	bar.finishCb = spb.wg.Done
	spb.bars = append(spb.bars, bar)

	spb.pool.Add(bar.ProgressBar)
	if start {
		spb.pool.Start()
	}
	return bar
}

// BasicProgressBar is packer's basic progress bar.
// Current implementation will always try to keep
// itself at the bottom of a terminal.
type BasicProgressBar struct {
	*pb.ProgressBar
	finishCb func()
}

func NewProgressBar(identifier string) *BasicProgressBar {
	bar := new(BasicProgressBar)
	bar.ProgressBar = pb.New(0)
	return bar
}

var _ ProgressBar = new(BasicProgressBar)

func (bpb *BasicProgressBar) Start(total int64) {
	bpb.SetTotal64(total)
	bpb.ProgressBar.Start()
}

func (bpb *BasicProgressBar) Finish() {
	if bpb.finishCb != nil {
		bpb.finishCb()
	}
	bpb.ProgressBar.Finish()
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
