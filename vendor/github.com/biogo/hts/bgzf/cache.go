// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bgzf

import (
	"bytes"
	"compress/gzip"
	"io"
)

// Cache is a Block caching type. Basic cache implementations are provided
// in the cache package. A Cache must be safe for concurrent use.
//
// If a Cache is a Wrapper, its Wrap method is called on newly created blocks.
type Cache interface {
	// Get returns the Block in the Cache with the specified
	// base or a nil Block if it does not exist. The returned
	// Block must be removed from the Cache.
	Get(base int64) Block

	// Put inserts a Block into the Cache, returning the Block
	// that was evicted or nil if no eviction was necessary and
	// a boolean indicating whether the put Block was retained
	// by the Cache.
	Put(Block) (evicted Block, retained bool)

	// Peek returns whether a Block exists in the cache for the
	// given base. If a Block satisfies the request, then exists
	// is returned as true with the offset for the next Block in
	// the stream, otherwise false and -1.
	Peek(base int64) (exists bool, next int64)
}

// Wrapper defines Cache types that need to modify a Block at its creation.
type Wrapper interface {
	Wrap(Block) Block
}

// Block wraps interaction with decompressed BGZF data blocks.
type Block interface {
	// Base returns the file offset of the start of
	// the gzip member from which the Block data was
	// decompressed.
	Base() int64

	io.Reader

	// Used returns whether one or more bytes have
	// been read from the Block.
	Used() bool

	// header returns the gzip.Header of the gzip member
	// from which the Block data was decompressed.
	header() gzip.Header

	// isMagicBlock returns whether the Block is a BGZF
	// magic EOF marker block.
	isMagicBlock() bool

	// ownedBy returns whether the Block is owned by
	// the given Reader.
	ownedBy(*Reader) bool

	// setOwner changes the owner to the given Reader,
	// reseting other data to its zero state.
	setOwner(*Reader)

	// hasData returns whether the Block has read data.
	hasData() bool

	// The following are unexported equivalents
	// of the io interfaces. seek is limited to
	// the file origin offset case and does not
	// return the new offset.
	seek(offset int64) error
	readFrom(io.ReadCloser) error

	// len returns the number of remaining
	// bytes that can be read from the Block.
	len() int

	// setBase sets the file offset of the start
	// and of the gzip member that the Block data
	// was decompressed from.
	setBase(int64)

	// NextBase returns the expected position of the next
	// BGZF block. It returns -1 if the Block is not valid.
	NextBase() int64

	// setHeader sets the file header of of the gzip
	// member that the Block data was decompressed from.
	setHeader(gzip.Header)

	// txOffset returns the current vitual offset.
	txOffset() Offset
}

type block struct {
	owner *Reader
	used  bool

	base  int64
	h     gzip.Header
	magic bool

	offset Offset

	buf  *bytes.Reader
	data [MaxBlockSize]byte
}

func (b *block) Base() int64 { return b.base }

func (b *block) Used() bool { return b.used }

func (b *block) Read(p []byte) (int, error) {
	n, err := b.buf.Read(p)
	b.offset.Block += uint16(n)
	if n > 0 {
		b.used = true
	}
	return n, err
}

func (b *block) readFrom(r io.ReadCloser) error {
	o := b.owner
	b.owner = nil
	buf := bytes.NewBuffer(b.data[:0])
	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}
	b.buf = bytes.NewReader(buf.Bytes())
	b.owner = o
	b.magic = b.magic && b.len() == 0
	return r.Close()
}

func (b *block) seek(offset int64) error {
	_, err := b.buf.Seek(offset, 0)
	if err == nil {
		b.offset.Block = uint16(offset)
	}
	return err
}

func (b *block) len() int {
	if b.buf == nil {
		return 0
	}
	return b.buf.Len()
}

func (b *block) setBase(n int64) {
	b.base = n
	b.offset = Offset{File: n}
}

func (b *block) NextBase() int64 {
	size := int64(expectedMemberSize(b.h))
	if size == -1 {
		return -1
	}
	return b.base + size
}

func (b *block) setHeader(h gzip.Header) {
	b.h = h
	b.magic = h.OS == 0xff &&
		h.ModTime.Equal(unixEpoch) &&
		h.Name == "" &&
		h.Comment == "" &&
		bytes.Equal(h.Extra, []byte("BC\x02\x00\x1b\x00"))
}

func (b *block) header() gzip.Header { return b.h }

func (b *block) isMagicBlock() bool { return b.magic }

func (b *block) setOwner(r *Reader) {
	b.owner = r
	b.used = false
	b.base = -1
	b.h = gzip.Header{}
	b.offset = Offset{}
	b.buf = nil
}

func (b *block) ownedBy(r *Reader) bool { return b.owner == r }

func (b *block) hasData() bool { return b.buf != nil }

func (b *block) txOffset() Offset { return b.offset }
