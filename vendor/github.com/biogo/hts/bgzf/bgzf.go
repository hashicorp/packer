// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bgzf implements BGZF format reading and writing according to the
// SAM specification.
//
// The specification is available at https://github.com/samtools/hts-specs.
package bgzf

import (
	"errors"
	"io"
	"os"
	"time"
)

const (
	BlockSize    = 0x0ff00 // The maximum size of an uncompressed input data block.
	MaxBlockSize = 0x10000 // The maximum size of a compressed output block.
)

const (
	bgzfExtra = "BC\x02\x00\x00\x00"
	minFrame  = 20 + len(bgzfExtra) // Minimum bgzf header+footer length.

	// Magic EOF block.
	magicBlock = "\x1f\x8b\x08\x04\x00\x00\x00\x00\x00\xff\x06\x00\x42\x43\x02\x00\x1b\x00\x03\x00\x00\x00\x00\x00\x00\x00\x00\x00"
)

var (
	bgzfExtraPrefix = []byte(bgzfExtra[:4])
	unixEpoch       = time.Unix(0, 0)
)

func compressBound(srcLen int) int {
	return srcLen + srcLen>>12 + srcLen>>14 + srcLen>>25 + 13 + minFrame
}

func init() {
	if compressBound(BlockSize) > MaxBlockSize {
		panic("bam: BlockSize too large")
	}
}

var (
	ErrClosed            = errors.New("bgzf: use of closed writer")
	ErrBlockOverflow     = errors.New("bgzf: block overflow")
	ErrWrongFileType     = errors.New("bgzf: file is a directory")
	ErrNoEnd             = errors.New("bgzf: cannot determine offset from end")
	ErrNotASeeker        = errors.New("bgzf: not a seeker")
	ErrContaminatedCache = errors.New("bgzf: cache owner mismatch")
	ErrNoBlockSize       = errors.New("bgzf: could not determine block size")
	ErrBlockSizeMismatch = errors.New("bgzf: unexpected block size")
)

// HasEOF checks for the presence of a BGZF magic EOF block.
// The magic block is defined in the SAM specification. A magic block
// is written by a Writer on calling Close. The ReaderAt must provide
// some method for determining valid ReadAt offsets.
func HasEOF(r io.ReaderAt) (bool, error) {
	type sizer interface {
		Size() int64
	}
	type stater interface {
		Stat() (os.FileInfo, error)
	}
	type lenSeeker interface {
		io.Seeker
		Len() int
	}
	var size int64
	switch r := r.(type) {
	case sizer:
		size = r.Size()
	case stater:
		fi, err := r.Stat()
		if err != nil {
			return false, err
		}
		size = fi.Size()
	case lenSeeker:
		var err error
		size, err = r.Seek(0, 1)
		if err != nil {
			return false, err
		}
		size += int64(r.Len())
	default:
		return false, ErrNoEnd
	}

	b := make([]byte, len(magicBlock))
	_, err := r.ReadAt(b, size-int64(len(magicBlock)))
	if err != nil {
		return false, err
	}
	for i, c := range b {
		if c != magicBlock[i] {
			return false, nil
		}
	}
	return true, nil
}
