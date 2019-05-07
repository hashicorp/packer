// +build !solaris

package packer

import (
	"io"
	"path/filepath"
	"sync"

	pb "github.com/cheggaaa/pb"
)

func ProgressBarConfig(bar *pb.ProgressBar, prefix string) {
	bar.SetUnits(pb.U_BYTES)
	bar.Prefix(prefix)
}

var defaultUiProgressBar = &uiProgressBar{}

// uiProgressBar is a self managed progress bar singleton.
// decorate your struct with a *uiProgressBar to
// give it TrackProgress capabilities.
// In TrackProgress if uiProgressBar is nil
// defaultUiProgressBar will be used as
// the progress bar.
type uiProgressBar struct {
	lock sync.Mutex

	pool *pb.Pool

	pbs int
}

func (p *uiProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	if p == nil {
		return defaultUiProgressBar.TrackProgress(src, currentSize, totalSize, stream)
	}
	p.lock.Lock()
	defer p.lock.Unlock()

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
