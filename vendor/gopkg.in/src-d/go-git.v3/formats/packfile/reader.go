package packfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"gopkg.in/src-d/go-git.v3/core"

	"github.com/klauspost/compress/zlib"
)

type Format int

var (
	EmptyRepositoryErr        = newError("empty repository")
	UnsupportedVersionErr     = newError("unsupported packfile version")
	MaxObjectsLimitReachedErr = newError("max. objects limit reached")
	MalformedPackfileErr      = newError("malformed pack file, does not start with 'PACK'")
	InvalidObjectErr          = newError("invalid git object")
	PatchingErr               = newError("patching error")
	PackEntryNotFoundErr      = newError("can't find a pack entry")
	ObjectNotFoundErr         = newError("can't find a object")
	ZLibErr                   = newError("zlib reading error")
)

const (
	DefaultMaxObjectsLimit = 1 << 20

	VersionSupported        = 2
	UnknownFormat    Format = 0
	OFSDeltaFormat   Format = 1
	REFDeltaFormat   Format = 2
)

// Reader reads a packfile from a binary string splitting it on objects
type Reader struct {
	// MaxObjectsLimit is the limit of objects to be load in the packfile, if
	// a packfile excess this number an error is throw, the default value
	// is defined by DefaultMaxObjectsLimit, usually the default limit is more
	// than enough to work with any repository, working extremly big repositories
	// where the number of object is bigger the memory can be exhausted.
	MaxObjectsLimit uint32

	// Format specifies if we are using ref-delta's or ofs-delta's, choosing the
	// correct format the memory usage is optimized
	// https://github.com/git/git/blob/8d530c4d64ffcc853889f7b385f554d53db375ed/Documentation/technical/protocol-capabilities.txt#L154
	Format Format

	r       *trackingReader
	s       core.ObjectStorage
	offsets map[int64]core.Hash
}

// NewReader returns a new Reader that reads from a io.Reader
func NewReader(r io.Reader) *Reader {
	return &Reader{
		MaxObjectsLimit: DefaultMaxObjectsLimit,

		r:       NewTrackingReader(r),
		offsets: make(map[int64]core.Hash, 0),
	}
}

// Read reads the objects and stores it at the ObjectStorage
func (r *Reader) Read(s core.ObjectStorage) (int64, error) {
	r.s = s
	if err := r.validateHeader(); err != nil {
		if err == io.EOF {
			return -1, EmptyRepositoryErr
		}

		return -1, err
	}

	version, err := r.readInt32()
	if err != nil {
		return -1, err
	}

	if version > VersionSupported {
		return -1, UnsupportedVersionErr
	}

	count, err := r.readInt32()
	if err != nil {
		return -1, err
	}

	if count > r.MaxObjectsLimit {
		return -1, MaxObjectsLimitReachedErr
	}

	return r.r.position, r.readObjects(count)
}

func (r *Reader) validateHeader() error {
	var header = make([]byte, 4)
	if _, err := io.ReadFull(r.r, header); err != nil {
		return err
	}

	if !bytes.Equal(header, []byte{'P', 'A', 'C', 'K'}) {
		return MalformedPackfileErr
	}

	return nil
}

func (r *Reader) readInt32() (uint32, error) {
	var value uint32
	if err := binary.Read(r.r, binary.BigEndian, &value); err != nil {
		return 0, err
	}

	return value, nil
}

func (r *Reader) readObjects(count uint32) error {
	// This code has 50-80 µs of overhead per object not counting zlib inflation.
	// Together with zlib inflation, it's 400-410 µs for small objects.
	// That's 1 sec for ~2450 objects, ~4.20 MB, or ~250 ms per MB,
	// of which 12-20 % is _not_ zlib inflation (ie. is our code).
	for i := 0; i < int(count); i++ {
		start := r.r.position
		obj, err := r.newObject()
		if err != nil && err != io.EOF {
			return err
		}

		if r.Format == UnknownFormat || r.Format == OFSDeltaFormat {
			r.offsets[start] = obj.Hash()
		}

		r.s.Set(obj)
		if err == io.EOF {
			break
		}
	}

	return nil
}

func (r *Reader) newObject() (core.Object, error) {
	raw, err := r.s.New()
	if err != nil {
		return nil, err
	}
	var steps int64

	var buf [1]byte
	if _, err := r.r.Read(buf[:]); err != nil {
		return nil, err
	}

	typ := core.ObjectType((buf[0] >> 4) & 7)
	size := int64(buf[0] & 15)
	steps++ // byte we just read to get `o.typ` and `o.size`

	var shift uint = 4
	for buf[0]&0x80 == 0x80 {
		if _, err := r.r.Read(buf[:]); err != nil {
			return nil, err
		}

		size += int64(buf[0]&0x7f) << shift
		steps++ // byte we just read to update `o.size`
		shift += 7
	}

	raw.SetType(typ)
	raw.SetSize(size)

	switch raw.Type() {
	case core.REFDeltaObject:
		err = r.readREFDelta(raw)
	case core.OFSDeltaObject:
		err = r.readOFSDelta(raw, steps)
	case core.CommitObject, core.TreeObject, core.BlobObject, core.TagObject:
		err = r.readObject(raw)
	default:
		err = InvalidObjectErr.n("tag %q", raw.Type)
	}

	return raw, err
}

func (r *Reader) readREFDelta(raw core.Object) (err error) {
	var ref core.Hash
	if _, err := io.ReadFull(r.r, ref[:]); err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	if err := r.inflate(buf); err != nil {
		return err
	}

	referenced, err := r.s.Get(ref)
	if err != nil {
		if err == core.ObjectNotFoundErr {
			return ObjectNotFoundErr.n("%s", ref)
		}
		return err
	}

	reader, err := referenced.Reader()
	if err != nil {
		return err
	}
	defer checkClose(reader, &err)

	d, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	patched := patchDelta(d, buf.Bytes())
	if patched == nil {
		return PatchingErr.n("hash %q", ref)
	}

	raw.SetType(referenced.Type())
	raw.SetSize(int64(len(patched)))

	writer, err := raw.Writer()
	if err != nil {
		return err
	}
	defer checkClose(writer, &err)

	writer.Write(patched)

	return nil
}

func (r *Reader) readOFSDelta(raw core.Object, steps int64) (err error) {
	start := r.r.position
	offset, err := decodeOffset(r.r, steps)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	if err = r.inflate(buf); err != nil {
		return err
	}

	ref, ok := r.offsets[start+offset]
	if !ok {
		return PackEntryNotFoundErr.n("offset %d", start+offset)
	}

	referenced, err := r.s.Get(ref)
	if err != nil {
		return err
	}

	reader, err := referenced.Reader()
	if err != nil {
		return err
	}
	defer checkClose(reader, &err)

	d, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	patched := patchDelta(d, buf.Bytes())
	if patched == nil {
		return PatchingErr.n("hash %q", ref)
	}

	raw.SetType(referenced.Type())
	raw.SetSize(int64(len(patched)))

	writer, err := raw.Writer()
	if err != nil {
		return err
	}
	defer checkClose(writer, &err)

	writer.Write(patched)

	return nil
}

func (r *Reader) readObject(raw core.Object) (err error) {
	writer, err := raw.Writer()
	if err != nil {
		return err
	}
	defer checkClose(writer, &err)

	return r.inflate(writer)
}

func (r *Reader) inflate(w io.Writer) error {
	zr, err := zlib.NewReader(r.r)
	if err != nil {
		if err == zlib.ErrHeader {
			return zlib.ErrHeader
		}

		return ZLibErr.n("%s", err)
	}

	defer zr.Close()

	_, err = io.Copy(w, zr)
	return err
}

type ReaderError struct {
	reason, additional string
}

func newError(reason string) *ReaderError {
	return &ReaderError{reason: reason}
}

func (e *ReaderError) Error() string {
	if e.additional == "" {
		return e.reason
	}

	return fmt.Sprintf("%s: %s", e.reason, e.additional)
}

func (e *ReaderError) n(format string, args ...interface{}) *ReaderError {
	return &ReaderError{
		reason:     e.reason,
		additional: fmt.Sprintf(format, args...),
	}
}
