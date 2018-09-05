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
	Start(total uint64)
	Add(current uint64)
	NewProxyReader(r io.Reader) (proxy io.Reader)
	Finish()
}

type StackableProgressBar struct {
	items   int32
	total   uint64
	started bool
	BasicProgressBar
	startOnce sync.Once
	group     sync.WaitGroup
}

var _ ProgressBar = new(StackableProgressBar)

func (spb *StackableProgressBar) start() {
	spb.BasicProgressBar.ProgressBar = pb.New(0)
	spb.BasicProgressBar.ProgressBar.SetUnits(pb.U_BYTES)

	spb.BasicProgressBar.ProgressBar.Start()
	go func() {
		spb.group.Wait()
		spb.BasicProgressBar.ProgressBar.Finish()
		spb.startOnce = sync.Once{}
		spb.BasicProgressBar.ProgressBar = nil
	}()
}

func (spb *StackableProgressBar) Start(total uint64) {
	atomic.AddUint64(&spb.total, total)
	atomic.AddInt32(&spb.items, 1)
	spb.group.Add(1)
	spb.startOnce.Do(spb.start)
	spb.SetTotal64(int64(atomic.LoadUint64(&spb.total)))
	spb.prefix()
}

func (spb *StackableProgressBar) prefix() {
	spb.BasicProgressBar.ProgressBar.Prefix(fmt.Sprintf("%d items: ", atomic.LoadInt32(&spb.items)))
}

func (spb *StackableProgressBar) Finish() {
	atomic.AddInt32(&spb.items, -1)
	spb.group.Done()
	spb.prefix()
}

type BasicProgressBar struct {
	*pb.ProgressBar
}

var _ ProgressBar = new(BasicProgressBar)

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

// NoopProgressBar is a silent progress bar
type NoopProgressBar struct {
}

var _ ProgressBar = new(NoopProgressBar)

func (npb *NoopProgressBar) Start(uint64)                                     {}
func (npb *NoopProgressBar) Add(uint64)                                       {}
func (npb *NoopProgressBar) Finish()                                          {}
func (npb *NoopProgressBar) NewProxyReader(r io.Reader) io.Reader             { return r }
func (npb *NoopProgressBar) NewProxyReadCloser(r io.ReadCloser) io.ReadCloser { return r }

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
