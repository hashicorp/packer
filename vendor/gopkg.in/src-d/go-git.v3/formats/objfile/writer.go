package objfile

import (
	"errors"
	"io"

	"gopkg.in/src-d/go-git.v3/core"

	"github.com/klauspost/compress/zlib"
)

var (
	// ErrOverflow is returned when an attempt is made to write more data than
	// was declared in NewWriter.
	ErrOverflow = errors.New("objfile: declared data length exceeded (overflow)")
)

// Writer writes and encodes data in compressed objfile format to a provided
// io.Writer.
//
// Writer implements io.WriteCloser. Close should be called when finished with
// the Writer. Close will not close the underlying io.Writer.
type Writer struct {
	header header
	hash   core.Hash // final computed hash stored after Close

	w          io.Writer      // provided writer wrapped in compressor and tee
	compressor io.WriteCloser // provided writer wrapped in compressor, retained for calling Close
	h          core.Hasher    // streaming SHA1 hash of encoded data
	written    int64          // Number of bytes written
}

// NewWriter returns a new Writer writing to w.
//
// The provided t is the type of object being written. The provided size is the
// number of uncompressed bytes being written.
//
// Calling NewWriter causes it to immediately write header data containing
// size and type information. Any errors encountered in that process will be
// returned in err.
//
// If an invalid t is provided, core.ErrInvalidType is returned. If a negative
// size is provided, ErrNegativeSize is returned.
//
// The returned Writer implements io.WriteCloser. Close should be called when
// finished with the Writer. Close will not close the underlying io.Writer.
func NewWriter(w io.Writer, t core.ObjectType, size int64) (*Writer, error) {
	if !t.Valid() {
		return nil, core.ErrInvalidType
	}
	if size < 0 {
		return nil, ErrNegativeSize
	}
	writer := &Writer{
		header: header{t: t, size: size},
	}
	return writer, writer.init(w)
}

// init prepares the zlib compressor for the given output as well as a hasher
// for computing its hash.
//
// init immediately writes header data to the output. This leaves the writer in
// a state that is ready to write content.
func (w *Writer) init(output io.Writer) (err error) {
	w.compressor = zlib.NewWriter(output)

	err = w.header.Write(w.compressor)
	if err != nil {
		w.compressor.Close()
		return
	}

	w.h = core.NewHasher(w.header.t, w.header.size)
	w.w = io.MultiWriter(w.compressor, w.h) // All writes to the compressor also write to the hash

	return
}

// Write reads len(p) from p to the object data stream. It returns the number of
// bytes written from p (0 <= n <= len(p)) and any error encountered that caused
// the write to stop early. The slice data contained in p will not be modified.
//
// If writing len(p) bytes would exceed the size provided in NewWriter,
// ErrOverflow is returned without writing any data.
func (w *Writer) Write(p []byte) (n int, err error) {
	if w.w == nil {
		return 0, ErrClosed
	}

	if w.written+int64(len(p)) > w.header.size {
		return 0, ErrOverflow
	}

	n, err = w.w.Write(p)
	w.written += int64(n)

	return
}

// Type returns the type of the object.
func (w *Writer) Type() core.ObjectType {
	return w.header.t
}

// Size returns the uncompressed size of the object in bytes.
func (w *Writer) Size() int64 {
	return w.header.size
}

// Hash returns the hash of the object data stream that has been written so far.
// It can be called before or after Close.
func (w *Writer) Hash() core.Hash {
	if w.w != nil {
		return w.h.Sum() // Not yet closed, return hash of data written so far
	}
	return w.hash
}

// Close releases any resources consumed by the Writer.
//
// Calling Close does not close the wrapped io.Writer originally passed to
// NewWriter.
func (w *Writer) Close() (err error) {
	if w.w == nil {
		// TODO: Consider returning ErrClosed here?
		return nil // Already closed
	}

	// Release the compressor's resources
	err = w.compressor.Close()

	// Save the hash because we're about to throw away the hasher
	w.hash = w.h.Sum()

	// Release references
	w.w = nil // Indicates closed state
	w.compressor = nil
	w.h.Hash = nil

	return
}
