// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bgzf

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

// Writer implements BGZF blocked gzip compression.
type Writer struct {
	gzip.Header
	w io.Writer

	active *compressor

	queue chan *compressor
	qwg   sync.WaitGroup

	waiting chan *compressor

	wg sync.WaitGroup

	closed bool

	m   sync.Mutex
	err error
}

// NewWriter returns a new Writer. Writes to the returned writer are
// compressed and written to w.
//
// The number of concurrent write compressors is specified by wc.
func NewWriter(w io.Writer, wc int) *Writer {
	bg, _ := NewWriterLevel(w, gzip.DefaultCompression, wc)
	return bg
}

// NewWriterLevel returns a new Writer using the specified compression level
// instead of gzip.DefaultCompression. Allowable level options are integer
// values between between gzip.BestSpeed and gzip.BestCompression inclusive.
//
// The number of concurrent write compressors is specified by wc.
func NewWriterLevel(w io.Writer, level, wc int) (*Writer, error) {
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		return nil, fmt.Errorf("bgzf: invalid compression level: %d", level)
	}
	wc++ // We count one for the active compressor.
	if wc < 2 {
		wc = 2
	}
	bg := &Writer{
		w:       w,
		waiting: make(chan *compressor, wc),
		queue:   make(chan *compressor, wc),
	}

	c := make([]compressor, wc)
	for i := range c {
		c[i].Header = &bg.Header
		c[i].level = level
		c[i].waiting = bg.waiting
		c[i].flush = make(chan *compressor, 1)
		c[i].qwg = &bg.qwg
		bg.waiting <- &c[i]
	}
	bg.active = <-bg.waiting

	bg.wg.Add(1)
	go func() {
		defer bg.wg.Done()
		for qw := range bg.queue {
			if !writeOK(bg, <-qw.flush) {
				break
			}
		}
	}()

	return bg, nil
}

func writeOK(bg *Writer, c *compressor) bool {
	defer func() { bg.waiting <- c }()

	if c.err != nil {
		bg.setErr(c.err)
		return false
	}
	if c.buf.Len() == 0 {
		return true
	}

	_, err := io.Copy(bg.w, &c.buf)
	bg.qwg.Done()
	if err != nil {
		bg.setErr(err)
		return false
	}
	c.next = 0

	return true
}

type compressor struct {
	*gzip.Header
	gz    *gzip.Writer
	level int

	next  int
	block [BlockSize]byte
	buf   bytes.Buffer

	flush chan *compressor
	qwg   *sync.WaitGroup

	waiting chan *compressor

	err error
}

func (c *compressor) writeBlock() {
	defer func() { c.flush <- c }()

	if c.gz == nil {
		c.gz, c.err = gzip.NewWriterLevel(&c.buf, c.level)
		if c.err != nil {
			return
		}
	} else {
		c.gz.Reset(&c.buf)
	}
	c.gz.Header = gzip.Header{
		Comment: c.Comment,
		Extra:   append([]byte(bgzfExtra), c.Extra...),
		ModTime: c.ModTime,
		Name:    c.Name,
		OS:      c.OS,
	}

	_, c.err = c.gz.Write(c.block[:c.next])
	if c.err != nil {
		return
	}
	c.err = c.gz.Close()
	if c.err != nil {
		return
	}
	c.next = 0

	b := c.buf.Bytes()
	i := bytes.Index(b, bgzfExtraPrefix)
	if i < 0 {
		c.err = gzip.ErrHeader
		return
	}
	size := len(b) - 1
	if size >= MaxBlockSize {
		c.err = ErrBlockOverflow
		return
	}
	b[i+4], b[i+5] = byte(size), byte(size>>8)
}

// Next returns the index of the start of the next write within the
// decompressed data block.
func (bg *Writer) Next() (int, error) {
	if bg.closed {
		return 0, ErrClosed
	}
	if err := bg.Error(); err != nil {
		return 0, err
	}

	return bg.active.next, nil
}

// Write writes the compressed form of b to the underlying io.Writer.
// Decompressed data blocks are limited to BlockSize, so individual
// byte slices may span block boundaries, however the Writer attempts
// to keep each write within a single data block.
func (bg *Writer) Write(b []byte) (int, error) {
	if bg.closed {
		return 0, ErrClosed
	}
	err := bg.Error()
	if err != nil {
		return 0, err
	}

	c := bg.active
	var n int
	for ; len(b) > 0 && err == nil; err = bg.Error() {
		var _n int
		if c.next == 0 || c.next+len(b) <= len(c.block) {
			_n = copy(c.block[c.next:], b)
			b = b[_n:]
			c.next += _n
			n += _n
		}

		if c.next == len(c.block) || _n == 0 {
			bg.queue <- c
			bg.qwg.Add(1)
			go c.writeBlock()
			c = <-bg.waiting
		}
	}
	bg.active = c

	return n, bg.Error()
}

// Flush writes unwritten data to the underlying io.Writer. Flush does not block.
func (bg *Writer) Flush() error {
	if bg.closed {
		return ErrClosed
	}
	if err := bg.Error(); err != nil {
		return err
	}

	if bg.active.next == 0 {
		return nil
	}

	var c *compressor
	c, bg.active = bg.active, <-bg.waiting
	bg.queue <- c
	bg.qwg.Add(1)
	go c.writeBlock()

	return bg.Error()
}

// Wait waits for all pending writes to complete and returns the subsequent
// error state of the Writer.
func (bg *Writer) Wait() error {
	if err := bg.Error(); err != nil {
		return err
	}
	bg.qwg.Wait()
	return bg.Error()
}

// Error returns the error state of the Writer.
func (bg *Writer) Error() error {
	bg.m.Lock()
	defer bg.m.Unlock()
	return bg.err
}

func (bg *Writer) setErr(err error) {
	bg.m.Lock()
	defer bg.m.Unlock()
	if bg.err == nil {
		bg.err = err
	}
}

// Close closes the Writer, waiting for any pending writes before returning
// the final error of the Writer.
func (bg *Writer) Close() error {
	if !bg.closed {
		c := bg.active
		bg.queue <- c
		bg.qwg.Add(1)
		<-bg.waiting
		c.writeBlock()
		bg.closed = true
		close(bg.queue)
		bg.wg.Wait()
		if bg.err == nil {
			_, bg.err = bg.w.Write([]byte(magicBlock))
		}
	}
	return bg.err
}
