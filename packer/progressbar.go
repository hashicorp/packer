// +build !solaris

package packer

import (
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	pb "github.com/cheggaaa/pb"
)

func ProgressBarConfig(bar *pb.ProgressBar, prefix string) {
	bar.SetUnits(pb.U_BYTES)
	bar.Prefix(prefix)
}

// UiProgressBar is a progress bar compatible with go-getter used in our
// UI structs.
type UiProgressBar struct {
	lock sync.Mutex
	pool *pb.Pool
	pbs  int
}

func (p *UiProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	if p == nil {

		return stream
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	// custom prefix used to track configured waits rather than file downloads
	// TODO: next time we feel we can justify a breaking interface change, this
	// deserves its own method on the progress bar rather than this hacked
	// workaround.
	if strings.Contains(src, "-packerwaiter-") {
		realPrefix := strings.Replace(src, "-packerwaiter-", "", -1)
		return p.TrackProgressWait(realPrefix, currentSize, totalSize, stream)
	}
	return p.TrackProgressFile(src, currentSize, totalSize, stream)

}

func (p *UiProgressBar) TrackProgressWait(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	newPb := pb.New64(totalSize)
	newPb.SetUnits(pb.U_DURATION)
	newPb.ShowPercent = false
	newPb.Prefix(src)

	if p.pool == nil {

		pool := pb.NewPool()
		err := pool.Start()
		if err != nil {
			// here, we probably cannot lock
			// stdout, so let's just return
			// stream to avoid any error.
			return stream
		}
		p.pool = pool
	}
	p.pool.Add(newPb)

	p.pbs++
	return &readCloser{
		Reader: nil,
		close: func() error {

			for i := currentSize; i < totalSize; i++ {
				newPb.Increment()
				time.Sleep(time.Second)
			}
			newPb.Finish()
			p.lock.Lock()
			defer p.lock.Unlock()
			p.pbs--
			if p.pbs <= 0 {
				p.pool.Stop()
				p.pool = nil
			}
			return nil
		},
	}
}

func (p *UiProgressBar) TrackProgressFile(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	newPb := pb.New64(totalSize)
	newPb.Set64(currentSize)
	ProgressBarConfig(newPb, filepath.Base(src))

	if p.pool == nil {
		pool := pb.NewPool()
		err := pool.Start()
		if err != nil {
			// here, we probably cannot lock
			// stdout, so let's just return
			// stream to avoid any error.
			return stream
		}
		p.pool = pool
	}
	p.pool.Add(newPb)
	reader := newPb.NewProxyReader(stream)

	p.pbs++
	return &readCloser{
		Reader: reader,
		close: func() error {
			p.lock.Lock()
			defer p.lock.Unlock()

			newPb.Finish()
			p.pbs--
			if p.pbs <= 0 {
				p.pool.Stop()
				p.pool = nil
			}
			return nil
		},
	}
}

type readCloser struct {
	io.Reader
	close func() error
}

func (c *readCloser) Close() error { return c.close() }
