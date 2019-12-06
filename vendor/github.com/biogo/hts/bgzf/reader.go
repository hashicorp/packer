// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bgzf

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"runtime"
	"sync"
)

// countReader wraps flate.Reader, adding support for querying current offset.
type countReader struct {
	// Underlying Reader.
	fr flate.Reader

	// Offset within the underlying reader.
	off int64
}

// newCountReader returns a new countReader.
func newCountReader(r io.Reader) *countReader {
	switch r := r.(type) {
	case *countReader:
		panic("bgzf: illegal use of internal type")
	case flate.Reader:
		return &countReader{fr: r}
	default:
		return &countReader{fr: bufio.NewReader(r)}
	}
}

// Read is required to satisfy flate.Reader.
func (r *countReader) Read(p []byte) (int, error) {
	n, err := r.fr.Read(p)
	r.off += int64(n)
	return n, err
}

// ReadByte is required to satisfy flate.Reader.
func (r *countReader) ReadByte() (byte, error) {
	b, err := r.fr.ReadByte()
	if err == nil {
		r.off++
	}
	return b, err
}

// offset returns the current offset in the underlying reader.
func (r *countReader) offset() int64 { return r.off }

// seek moves the countReader to the specified offset using rs as the
// underlying reader.
func (r *countReader) seek(rs io.ReadSeeker, off int64) error {
	_, err := rs.Seek(off, 0)
	if err != nil {
		return err
	}

	type reseter interface {
		Reset(io.Reader)
	}
	switch cr := r.fr.(type) {
	case reseter:
		cr.Reset(rs)
	default:
		r.fr = newCountReader(rs)
	}
	r.off = off

	return nil
}

// buffer is a flate.Reader used by a decompressor to store read-ahead data.
type buffer struct {
	// Buffered compressed data from read ahead.
	off  int // Current position in buffered data.
	size int // Total size of buffered data.
	data [MaxBlockSize]byte
}

// Read provides the flate.Decompressor Read method.
func (r *buffer) Read(b []byte) (int, error) {
	if r.off >= r.size {
		return 0, io.EOF
	}
	if n := r.size - r.off; len(b) > n {
		b = b[:n]
	}
	n := copy(b, r.data[r.off:])
	r.off += n
	return n, nil
}

// ReadByte provides the flate.Decompressor ReadByte method.
func (r *buffer) ReadByte() (byte, error) {
	if r.off == r.size {
		return 0, io.EOF
	}
	b := r.data[r.off]
	r.off++
	return b, nil
}

// reset makes the buffer available to store data.
func (r *buffer) reset() { r.size = 0 }

// hasData returns whether the buffer has any data buffered.
func (r *buffer) hasData() bool { return r.size != 0 }

// readLimited reads n bytes into the buffer from the given source.
func (r *buffer) readLimited(n int, src *countReader) error {
	if r.hasData() {
		panic("bgzf: read into non-empty buffer")
	}
	r.off = 0
	var err error
	r.size, err = io.ReadFull(src, r.data[:n])
	return err
}

// equals returns a boolean indicating the equality between
// the buffered data and the given byte slice.
func (r *buffer) equals(b []byte) bool { return bytes.Equal(r.data[:r.size], b) }

// decompressor is a gzip member decompressor worker.
type decompressor struct {
	owner *Reader

	gz gzip.Reader

	cr *countReader

	// Current block size.
	blockSize int

	// Buffered compressed data from read ahead.
	buf buffer

	// Decompressed data.
	wg  sync.WaitGroup
	blk Block

	err error
}

// Read provides the Read method for the decompressor's gzip.Reader.
func (d *decompressor) Read(b []byte) (int, error) {
	if d.buf.hasData() {
		return d.buf.Read(b)
	}
	return d.cr.Read(b)
}

// ReadByte provides the ReadByte method for the decompressor's gzip.Reader.
func (d *decompressor) ReadByte() (byte, error) {
	if d.buf.hasData() {
		return d.buf.ReadByte()
	}
	return d.cr.ReadByte()
}

// lazyBlock conditionally creates a ready to use Block.
func (d *decompressor) lazyBlock() {
	if d.blk == nil {
		if w, ok := d.owner.cache.(Wrapper); ok {
			d.blk = w.Wrap(&block{owner: d.owner})
		} else {
			d.blk = &block{owner: d.owner}
		}
		return
	}
	if !d.blk.ownedBy(d.owner) {
		d.blk.setOwner(d.owner)
	}
}

// acquireHead gains the read head from the decompressor's owner.
func (d *decompressor) acquireHead() {
	d.wg.Add(1)
	d.cr = <-d.owner.head
}

// releaseHead releases the read head back to the decompressor's owner.
func (d *decompressor) releaseHead() {
	d.owner.head <- d.cr
	d.cr = nil // Defensively zero the reader.
}

// wait waits for the current member to be decompressed or fail, and returns
// the resulting error state.
func (d *decompressor) wait() (Block, error) {
	d.wg.Wait()
	blk := d.blk
	d.blk = nil
	return blk, d.err
}

// using sets the Block for the decompressor to work with.
func (d *decompressor) using(b Block) *decompressor { d.blk = b; return d }

// nextBlockAt makes the decompressor ready for reading decompressed data
// from its Block. It checks if there is a cached Block for the nextBase,
// otherwise it seeks to the correct location if decompressor is not
// correctly positioned, and then reads the compressed data and fills
// the decompressed Block.
// After nextBlockAt returns without error, the decompressor's Block
// holds a valid gzip.Header and base offset.
func (d *decompressor) nextBlockAt(off int64, rs io.ReadSeeker) *decompressor {
	d.err = nil
	for {
		exists, next := d.owner.cacheHasBlockFor(off)
		if !exists {
			break
		}
		off = next
	}

	d.lazyBlock()

	d.acquireHead()
	defer d.releaseHead()

	if d.cr.offset() != off {
		if rs == nil {
			// It should not be possible for the expected next block base
			// to be out of register with the count reader unless Seek
			// has been called, so we know the base reader must be an
			// io.ReadSeeker.
			var ok bool
			rs, ok = d.owner.r.(io.ReadSeeker)
			if !ok {
				panic("bgzf: unexpected offset without seek")
			}
		}
		d.err = d.cr.seek(rs, off)
		if d.err != nil {
			d.wg.Done()
			return d
		}
	}

	d.blk.setBase(d.cr.offset())
	d.err = d.readMember()
	if d.err != nil {
		d.wg.Done()
		return d
	}
	d.blk.setHeader(d.gz.Header)
	d.gz.Header = gzip.Header{} // Prevent retention of header field in next use.

	// Decompress data into the decompressor's Block.
	go func() {
		d.err = d.blk.readFrom(&d.gz)
		d.wg.Done()
	}()

	return d
}

// expectedMemberSize returns the size of the BGZF conformant gzip member.
// It returns -1 if no BGZF block size field is found.
func expectedMemberSize(h gzip.Header) int {
	i := bytes.Index(h.Extra, bgzfExtraPrefix)
	if i < 0 || i+5 >= len(h.Extra) {
		return -1
	}
	return (int(h.Extra[i+4]) | int(h.Extra[i+5])<<8) + 1
}

// readMember buffers the gzip member starting the current decompressor offset.
func (d *decompressor) readMember() error {
	// Set the decompressor to Read from the underlying flate.Reader
	// and mark the starting offset from which the underlying reader
	// was used.
	d.buf.reset()
	mark := d.cr.offset()

	err := d.gz.Reset(d)
	if err != nil {
		d.blockSize = -1
		return err
	}

	d.blockSize = expectedMemberSize(d.gz.Header)
	if d.blockSize < 0 {
		return ErrNoBlockSize
	}
	skipped := int(d.cr.offset() - mark)

	// Read compressed data into the decompressor buffer until the
	// underlying flate.Reader is positioned at the end of the gzip
	// member in which the readMember call was made.
	return d.buf.readLimited(d.blockSize-skipped, d.cr)
}

// Offset is a BGZF virtual offset.
type Offset struct {
	File  int64
	Block uint16
}

// Chunk is a region of a BGZF file.
type Chunk struct {
	Begin Offset
	End   Offset
}

// Reader implements BGZF blocked gzip decompression.
type Reader struct {
	gzip.Header
	r io.Reader

	// head serialises access to the underlying
	// io.Reader.
	head chan *countReader

	// lastChunk is the virtual file offset
	// interval of the last successful read
	// or seek operation.
	lastChunk Chunk

	// Blocked specifies the behaviour of the
	// Reader at the end of a BGZF member.
	// If the Reader is Blocked, a Read that
	// reaches the end of a BGZF block will
	// return io.EOF. This error is not sticky,
	// so a subsequent Read will progress to
	// the next block if it is available.
	Blocked bool

	// Non-concurrent work decompressor.
	dec *decompressor

	// Concurrent work fields.
	waiting chan *decompressor
	working chan *decompressor
	control chan int64

	current Block

	// cache is the Reader block cache. If Cache is not nil,
	// the cache is queried for blocks before an attempt to
	// read from the underlying io.Reader.
	mu    sync.RWMutex
	cache Cache

	err error
}

// NewReader returns a new BGZF reader.
//
// The number of concurrent read decompressors is specified by rd.
// If rd is 0, GOMAXPROCS concurrent will be created. The returned
// Reader should be closed after use to avoid leaking resources.
func NewReader(r io.Reader, rd int) (*Reader, error) {
	if rd == 0 {
		rd = runtime.GOMAXPROCS(0)
	}
	bg := &Reader{
		r: r,

		head: make(chan *countReader, 1),
	}
	bg.head <- newCountReader(r)

	// Make work loop control structures.
	if rd > 1 {
		bg.waiting = make(chan *decompressor, rd)
		bg.working = make(chan *decompressor, rd)
		bg.control = make(chan int64, 1)
		for ; rd > 1; rd-- {
			bg.waiting <- &decompressor{owner: bg}
		}
	}

	// Read the first block now so we can fail before
	// the first Read call if there is a problem.
	bg.dec = &decompressor{owner: bg}
	blk, err := bg.dec.nextBlockAt(0, nil).wait()
	if err != nil {
		return nil, err
	}
	bg.current = blk
	bg.Header = bg.current.header()

	// Set up work loop if rd was > 1.
	if bg.control != nil {
		bg.waiting <- bg.dec
		bg.dec = nil
		next := blk.NextBase()
		go func() {
			defer func() {
				bg.mu.Lock()
				bg.cache = nil
				bg.mu.Unlock()
			}()
			for dec := range bg.waiting {
				var open bool
				if next < 0 {
					next, open = <-bg.control
					if !open {
						return
					}
				} else {
					select {
					case next, open = <-bg.control:
						if !open {
							return
						}
					default:
					}
				}
				dec.nextBlockAt(next, nil)
				next = dec.blk.NextBase()
				bg.working <- dec
			}
		}()
	}

	return bg, nil
}

// SetCache sets the cache to be used by the Reader.
func (bg *Reader) SetCache(c Cache) {
	bg.mu.Lock()
	bg.cache = c
	bg.mu.Unlock()
}

// Seek performs a seek operation to the given virtual offset.
func (bg *Reader) Seek(off Offset) error {
	rs, ok := bg.r.(io.ReadSeeker)
	if !ok {
		return ErrNotASeeker
	}

	if off.File != bg.current.Base() || !bg.current.hasData() {
		ok := bg.cacheSwap(off.File)
		if !ok {
			var dec *decompressor
			if bg.dec != nil {
				dec = bg.dec
			} else {
				select {
				case dec = <-bg.waiting:
				case dec = <-bg.working:
					blk, err := dec.wait()
					if err == nil {
						bg.keep(blk)
					}
				}
			}
			bg.current, bg.err = dec.
				using(bg.current).
				nextBlockAt(off.File, rs).
				wait()
			if bg.dec == nil {
				select {
				case <-bg.control:
				default:
				}
				bg.control <- bg.current.NextBase()
				bg.waiting <- dec
			}
			bg.Header = bg.current.header()
			if bg.err != nil {
				return bg.err
			}
		}
	}

	bg.err = bg.current.seek(int64(off.Block))
	if bg.err == nil {
		bg.lastChunk = Chunk{Begin: off, End: off}
	}

	return bg.err
}

// LastChunk returns the region of the BGZF file read by the last read
// operation or the resulting virtual offset of the last successful
// seek operation.
func (bg *Reader) LastChunk() Chunk { return bg.lastChunk }

// BlockLen returns the number of bytes remaining to be read from the
// current BGZF block.
func (bg *Reader) BlockLen() int { return bg.current.len() }

// Close closes the reader and releases resources.
func (bg *Reader) Close() error {
	if bg.control != nil {
		close(bg.control)
		close(bg.waiting)
	}
	if bg.err == io.EOF {
		return nil
	}
	return bg.err
}

// Read implements the io.Reader interface.
func (bg *Reader) Read(p []byte) (int, error) {
	if bg.err != nil {
		return 0, bg.err
	}

	// Discard leading empty blocks. This is an indexing
	// optimisation to avoid retaining useless members
	// in a BAI/CSI.
	for bg.current.len() == 0 {
		bg.err = bg.nextBlock()
		if bg.err != nil {
			return 0, bg.err
		}
	}

	bg.lastChunk.Begin = bg.current.txOffset()

	var n int
	for n < len(p) && bg.err == nil {
		var _n int
		_n, bg.err = bg.current.Read(p[n:])
		n += _n
		if bg.err == io.EOF {
			if n == len(p) {
				bg.err = nil
				break
			}

			if bg.Blocked {
				bg.err = nil
				bg.lastChunk.End = bg.current.txOffset()
				return n, io.EOF
			}

			bg.err = bg.nextBlock()
			if bg.err != nil {
				break
			}
		}
	}

	bg.lastChunk.End = bg.current.txOffset()
	return n, bg.err
}

// nextBlock swaps the current decompressed block for the next
// in the stream. If the block is available from the cache
// no additional work is done, otherwise a decompressor is
// used or waited on.
func (bg *Reader) nextBlock() error {
	base := bg.current.NextBase()
	ok := bg.cacheSwap(base)
	if ok {
		bg.Header = bg.current.header()
		return nil
	}

	var err error
	if bg.dec != nil {
		bg.dec.using(bg.current).nextBlockAt(base, nil)
		bg.current, err = bg.dec.wait()
	} else {
		var ok bool
		for i := 0; i < cap(bg.working); i++ {
			dec := <-bg.working
			bg.current, err = dec.wait()
			bg.waiting <- dec
			if bg.current.Base() == base {
				ok = true
				break
			}
			if err == nil {
				bg.keep(bg.current)
				bg.current = nil
			}
		}
		if !ok {
			panic("bgzf: unexpected block")
		}
	}
	if err != nil {
		return err
	}

	// Only set header if there was no error.
	h := bg.current.header()
	if bg.current.isMagicBlock() {
		// TODO(kortschak): Do this more carefully. It may be that
		// someone actually has extra data in this field that we are
		// clobbering.
		bg.Header.Extra = h.Extra
	} else {
		bg.Header = h
	}

	return nil
}

// cacheSwap attempts to swap the current Block for a cached Block
// for the given base offset. It returns true if successful.
func (bg *Reader) cacheSwap(base int64) bool {
	bg.mu.RLock()
	defer bg.mu.RUnlock()
	if bg.cache == nil {
		return false
	}

	blk, err := bg.cachedBlockFor(base)
	if err != nil {
		return false
	}
	if blk != nil {
		// TODO(kortschak): Under some conditions, e.g. FIFO
		// cache we will be discarding a non-nil evicted Block.
		// Consider retaining these in a sync.Pool.
		bg.cachePut(bg.current)
		bg.current = blk
		return true
	}
	var retained bool
	bg.current, retained = bg.cachePut(bg.current)
	if retained {
		bg.current = nil
	}
	return false
}

// cacheHasBlockFor returns whether the Reader's cache has a block
// for the given base offset. If the requested Block exists, the base
// offset of the following Block is returned.
func (bg *Reader) cacheHasBlockFor(base int64) (exists bool, next int64) {
	bg.mu.RLock()
	defer bg.mu.RUnlock()
	if bg.cache == nil {
		return false, -1
	}
	return bg.cache.Peek(base)
}

// cachedBlockFor returns a non-nil Block if the Reader has access to a
// cache and the cache holds the block with the given base and the
// correct owner, otherwise it returns nil. If the Block's owner is not
// correct, or the Block cannot seek to the start of its data, a non-nil
// error is returned.
func (bg *Reader) cachedBlockFor(base int64) (Block, error) {
	blk := bg.cache.Get(base)
	if blk != nil {
		if !blk.ownedBy(bg) {
			return nil, ErrContaminatedCache
		}
		err := blk.seek(0)
		if err != nil {
			return nil, err
		}
	}
	return blk, nil
}

// cachePut puts the given Block into the cache if it exists, it returns
// the Block that was evicted or b if it was not retained, and whether
// the Block was retained by the cache.
func (bg *Reader) cachePut(b Block) (evicted Block, retained bool) {
	if b == nil || !b.hasData() {
		return b, false
	}
	return bg.cache.Put(b)
}

// keep puts the given Block into the cache if it exists.
func (bg *Reader) keep(b Block) {
	if b == nil || !b.hasData() {
		return
	}
	bg.mu.RLock()
	defer bg.mu.RUnlock()
	if bg.cache != nil {
		bg.cache.Put(b)
	}
}

// Begin returns a Tx that starts at the current virtual offset.
func (bg *Reader) Begin() Tx { return Tx{begin: bg.lastChunk.Begin, r: bg} }

// Tx represents a multi-read transaction.
type Tx struct {
	begin Offset
	r     *Reader
}

// End returns the Chunk spanning the transaction. After return the Tx is
// no longer valid.
func (t *Tx) End() Chunk {
	c := Chunk{Begin: t.begin, End: t.r.lastChunk.End}
	t.r = nil
	return c
}
