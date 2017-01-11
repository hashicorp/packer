package ansible

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/packer/packer"
)

const (
	scpOK         = "\x00"
	scpEmptyError = "\x02\n"
)

/*
scp is a simple, but poorly documented, protocol. Thankfully, its source is
freely available, and there is at least one page that describes it reasonably
well.

* https://raw.githubusercontent.com/openssh/openssh-portable/master/scp.c
* https://opensource.apple.com/source/OpenSSH/OpenSSH-7.1/openssh/scp.c
* https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works is a great
	resource, but has some bad information. Its first problem is that it doesn't
	correctly describe why the producer has to read more responses than messages
	it sends (because it has to read the 0 sent by the sink to start the
	transfer). The second problem is that it omits that the producer needs to
	send a 0 byte after file contents.
*/

func scpUploadSession(opts []byte, rest string, in io.Reader, out io.Writer, comm packer.Communicator) error {
	rest = strings.TrimSpace(rest)
	if len(rest) == 0 {
		fmt.Fprintf(out, scpEmptyError)
		return errors.New("no scp target specified")
	}

	d, err := ioutil.TempDir("", "packer-ansible-upload")
	if err != nil {
		fmt.Fprintf(out, scpEmptyError)
		return err
	}
	defer os.RemoveAll(d)

	// To properly implement scp, rest should be checked to see if it is a
	// directory on the remote side, but ansible only sends files, so there's no
	// need to set targetIsDir, because it can be safely assumed that rest is
	// intended to be a file, and whatever names are used in 'C' commands are
	// irrelavant.
	state := &scpUploadState{target: rest, srcRoot: d, comm: comm}

	fmt.Fprintf(out, scpOK) // signal the client to start the transfer.
	return state.Protocol(bufio.NewReader(in), out)
}

func scpDownloadSession(opts []byte, rest string, in io.Reader, out io.Writer, comm packer.Communicator) error {
	rest = strings.TrimSpace(rest)
	if len(rest) == 0 {
		fmt.Fprintf(out, scpEmptyError)
		return errors.New("no scp source specified")
	}

	d, err := ioutil.TempDir("", "packer-ansible-download")
	if err != nil {
		fmt.Fprintf(out, scpEmptyError)
		return err
	}
	defer os.RemoveAll(d)

	if bytes.Contains([]byte{'d'}, opts) {
		// the only ansible module that supports downloading via scp is fetch,
		// fetch only supports file downloads as of Ansible 2.1.
		fmt.Fprintf(out, scpEmptyError)
		return errors.New("directory downloads not supported")
	}

	f, err := os.Create(filepath.Join(d, filepath.Base(rest)))
	if err != nil {
		fmt.Fprintf(out, scpEmptyError)
		return err
	}
	defer f.Close()

	err = comm.Download(rest, f)
	if err != nil {
		fmt.Fprintf(out, scpEmptyError)
		return err
	}

	state := &scpDownloadState{srcRoot: d}

	return state.Protocol(bufio.NewReader(in), out)
}

func (state *scpDownloadState) FileProtocol(path string, info os.FileInfo, in *bufio.Reader, out io.Writer) error {
	size := info.Size()
	perms := fmt.Sprintf("C%04o", info.Mode().Perm())
	fmt.Fprintln(out, perms, size, info.Name())
	err := scpResponse(in)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	io.CopyN(out, f, size)
	fmt.Fprintf(out, scpOK)

	return scpResponse(in)
}

type scpUploadState struct {
	comm        packer.Communicator
	target      string // target is the directory on the target
	srcRoot     string // srcRoot is the directory on the host
	mtime       time.Time
	atime       time.Time
	dir         string // dir is a path relative to the roots
	targetIsDir bool
}

func (scp scpUploadState) DestPath() string {
	return filepath.Join(scp.target, scp.dir)
}

func (scp scpUploadState) SrcPath() string {
	return filepath.Join(scp.srcRoot, scp.dir)
}

func (state *scpUploadState) Protocol(in *bufio.Reader, out io.Writer) error {
	for {
		b, err := in.ReadByte()
		if err != nil {
			return err
		}
		switch b {
		case 'T':
			err := state.TimeProtocol(in, out)
			if err != nil {
				return err
			}
		case 'C':
			return state.FileProtocol(in, out)
		case 'E':
			state.dir = filepath.Dir(state.dir)
			fmt.Fprintf(out, scpOK)
			return nil
		case 'D':
			return state.DirProtocol(in, out)
		default:
			fmt.Fprintf(out, scpEmptyError)
			return fmt.Errorf("unexpected message: %c", b)
		}
	}
}

func (state *scpUploadState) FileProtocol(in *bufio.Reader, out io.Writer) error {
	defer func() {
		state.mtime = time.Time{}
	}()

	var mode os.FileMode
	var size int64
	var name string
	_, err := fmt.Fscanf(in, "%04o %d %s\n", &mode, &size, &name)
	if err != nil {
		fmt.Fprintf(out, scpEmptyError)
		return fmt.Errorf("invalid file message: %v", err)
	}
	fmt.Fprintf(out, scpOK)

	var fi os.FileInfo = fileInfo{name: name, size: size, mode: mode, mtime: state.mtime}

	dest := state.DestPath()
	if state.targetIsDir {
		dest = filepath.Join(dest, fi.Name())
	}

	err = state.comm.Upload(dest, io.LimitReader(in, fi.Size()), &fi)
	if err != nil {
		fmt.Fprintf(out, scpEmptyError)
		return err
	}

	err = scpResponse(in)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, scpOK)
	return nil
}

func (state *scpUploadState) TimeProtocol(in *bufio.Reader, out io.Writer) error {
	var m, a int64
	if _, err := fmt.Fscanf(in, "%d 0 %d 0\n", &m, &a); err != nil {
		fmt.Fprintf(out, scpEmptyError)
		return err
	}
	fmt.Fprintf(out, scpOK)

	state.atime = time.Unix(a, 0)
	state.mtime = time.Unix(m, 0)
	return nil
}

func (state *scpUploadState) DirProtocol(in *bufio.Reader, out io.Writer) error {
	var mode os.FileMode
	var length uint
	var name string

	if _, err := fmt.Fscanf(in, "%04o %d %s\n", &mode, &length, &name); err != nil {
		fmt.Fprintf(out, scpEmptyError)
		return fmt.Errorf("invalid directory message: %v", err)
	}
	fmt.Fprintf(out, scpOK)

	path := filepath.Join(state.dir, name)
	if err := os.Mkdir(path, mode); err != nil {
		return err
	}
	state.dir = path

	if state.atime.IsZero() {
		state.atime = time.Now()
	}
	if state.mtime.IsZero() {
		state.mtime = time.Now()
	}

	if err := os.Chtimes(path, state.atime, state.mtime); err != nil {
		return err
	}

	if err := state.comm.UploadDir(filepath.Dir(state.DestPath()), state.SrcPath(), nil); err != nil {
		return err
	}

	state.mtime = time.Time{}
	state.atime = time.Time{}
	return state.Protocol(in, out)
}

type scpDownloadState struct {
	srcRoot string // srcRoot is the directory on the host
}

func (state *scpDownloadState) Protocol(in *bufio.Reader, out io.Writer) error {
	r := bufio.NewReader(in)
	// read the byte sent by the other side to start the transfer
	scpResponse(r)

	return filepath.Walk(state.srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == state.srcRoot {
			return nil
		}

		if info.IsDir() {
			// no need to get fancy; srcRoot should only contain one file, because
			// Ansible only allows fetching a single file.
			return errors.New("unexpected directory")
		}

		return state.FileProtocol(path, info, r, out)
	})
}

func scpOptions(s string) (opts []byte, rest string) {
	end := 0
	opt := false

Loop:
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch {
		case b == ' ':
			opt = false
			end++
		case b == '-':
			opt = true
			end++
		case opt:
			opts = append(opts, b)
			end++
		default:
			break Loop
		}
	}

	rest = s[end:]
	return
}

func scpResponse(r *bufio.Reader) error {
	code, err := r.ReadByte()
	if err != nil {
		return err
	}

	if code != 0 {
		message, err := r.ReadString('\n')
		if err != nil {
			return fmt.Errorf("Error reading error message: %s", err)
		}

		// 1 is a warning. Anything higher (really just 2) is an error.
		if code > 1 {
			return errors.New(string(message))
		}

		log.Println("WARNING:", err)
	}
	return nil
}

type fileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	mtime time.Time
}

func (fi fileInfo) Name() string      { return fi.name }
func (fi fileInfo) Size() int64       { return fi.size }
func (fi fileInfo) Mode() os.FileMode { return fi.mode }
func (fi fileInfo) ModTime() time.Time {
	if fi.mtime.IsZero() {
		return time.Now()
	}
	return fi.mtime
}
func (fi fileInfo) IsDir() bool      { return fi.mode.IsDir() }
func (fi fileInfo) Sys() interface{} { return nil }
