package objfile

import (
	"errors"
	"io"

	"gopkg.in/src-d/go-git.v3/core"

	"github.com/klauspost/compress/zlib"
)

var (
	// ErrZLib is returned when the objfile contains invalid zlib data.
	ErrZLib = errors.New("objfile: invalid zlib data")
)

// Reader reads and decodes compressed objfile data from a provided io.Reader.
//
// Reader implements io.ReadCloser. Close should be called when finished with
// the Reader. Close will not close the underlying io.Reader.
type Reader struct {
	header header
	hash   core.Hash // final computed hash stored after Close

	r            io.Reader     // provided reader wrapped in decompressor and tee
	decompressor io.ReadCloser // provided reader wrapped in decompressor, retained for calling Close
	h            core.Hasher   // streaming SHA1 hash of decoded data
}

// NewReader returns a new Reader reading from r.
//
// Calling NewReader causes it to immediately read in header data from r
// containing size and type information. Any errors encountered in that
// process will be returned in err.
//
// The returned Reader implements io.ReadCloser. Close should be called when
// finished with the Reader. Close will not close the underlying io.Reader.
func NewReader(r io.Reader) (*Reader, error) {
	reader := &Reader{}
	return reader, reader.init(r)
}

// init prepares the zlib decompressor for the given input as well as a hasher
// for computing its hash.
//
// init immediately reads header data from the input and stores it. This leaves
// the Reader in a state that is ready to read content.
func (r *Reader) init(input io.Reader) (err error) {
	r.decompressor, err = zlib.NewReader(input)
	if err != nil {
		// TODO: Make this error match the ZLibErr in formats/packfile/reader.go?
		return ErrZLib
	}

	err = r.header.Read(r.decompressor)
	if err != nil {
		r.decompressor.Close()
		return
	}

	r.h = core.NewHasher(r.header.t, r.header.size)
	r.r = io.TeeReader(r.decompressor, r.h) // All reads from the decompressor also write to the hash

	return
}

// Read reads len(p) bytes into p from the object data stream. It returns
// the number of bytes read (0 <= n <= len(p)) and any error encountered. Even
// if Read returns n < len(p), it may use all of p as scratch space during the
// call.
//
// If Read encounters the end of the data stream it will return err == io.EOF,
// either in the current call if n > 0 or in a subsequent call.
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.r == nil {
		return 0, ErrClosed
	}

	return r.r.Read(p)
}

// Type returns the type of the object.
func (r *Reader) Type() core.ObjectType {
	return r.header.t
}

// Size returns the uncompressed size of the object in bytes.
func (r *Reader) Size() int64 {
	return r.header.size
}

// Hash returns the hash of the object data stream that has been read so far.
// It can be called before or after Close.
func (r *Reader) Hash() core.Hash {
	if r.r != nil {
		return r.h.Sum() // Not yet closed, return hash of data read so far
	}
	return r.hash
}

// Close releases any resources consumed by the Reader.
//
// Calling Close does not close the wrapped io.Reader originally passed to
// NewReader.
func (r *Reader) Close() (err error) {
	if r.r == nil {
		// TODO: Consider returning ErrClosed here?
		return nil // Already closed
	}

	// Release the decompressor's resources
	err = r.decompressor.Close()

	// Save the hash because we're about to throw away the hasher
	r.hash = r.h.Sum()

	// Release references
	r.r = nil // Indicates closed state
	r.decompressor = nil
	r.h.Hash = nil

	return
}
