// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !solaris
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
