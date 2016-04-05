package common

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"net/url"
	"os"
	"runtime"
	"path"
	"path/filepath"
	"strings"
)

import (
	"net/http"
	"github.com/jlaffeye/ftp"
	"bufio"
)


// DownloadConfig is the configuration given to instantiate a new
// download instance. Once a configuration is used to instantiate
// a download client, it must not be modified.
type DownloadConfig struct {
	// The source URL in the form of a string.
	Url string

	// This is the path to download the file to.
	TargetPath string

	// DownloaderMap maps a schema to a Download.
	DownloaderMap map[string]Downloader

	// If true, this will copy even a local file to the target
	// location. If false, then it will "download" the file by just
	// returning the local path to the file.
	CopyFile bool

	// The hashing implementation to use to checksum the downloaded file.
	Hash hash.Hash

	// The checksum for the downloaded file. The hash implementation configuration
	// for the downloader will be used to verify with this checksum after
	// it is downloaded.
	Checksum []byte

	// What to use for the user agent for HTTP requests. If set to "", use the
	// default user agent provided by Go.
	UserAgent string
}

// A DownloadClient helps download, verify checksums, etc.
type DownloadClient struct {
	config     *DownloadConfig
	downloader Downloader
}

// HashForType returns the Hash implementation for the given string
// type, or nil if the type is not supported.
func HashForType(t string) hash.Hash {
	switch t {
	case "md5":
		return md5.New()
	case "sha1":
		return sha1.New()
	case "sha256":
		return sha256.New()
	case "sha512":
		return sha512.New()
	default:
		return nil
	}
}

// NewDownloadClient returns a new DownloadClient for the given
// configuration.
func NewDownloadClient(c *DownloadConfig) *DownloadClient {
	const mtu = 1500 /* ethernet */ - 20 /* ipv4 */ - 20 /* tcp */

	if c.DownloaderMap == nil {
		c.DownloaderMap = map[string]Downloader{
			"file":  &FileDownloader{bufferSize: nil},
			"ftp":   &FTPDownloader{userInfo: url.Userinfo{username:"anonymous", password: "anonymous@"}, mtu: mtu},
			"http":  &HTTPDownloader{userAgent: c.UserAgent},
			"https": &HTTPDownloader{userAgent: c.UserAgent},
			"smb":   &SMBDownloader{bufferSize: nil}
		}
	}

	return &DownloadClient{config: c}
}

// A downloader is responsible for actually taking a remote URL and
// downloading it.
type Downloader interface {
	Cancel()
	Download(*os.File, *url.URL) error
	Progress() uint
	Total() uint
}

func (d *DownloadClient) Cancel() {
	// TODO(mitchellh): Implement
}

func (d *DownloadClient) Get() (string, error) {
	// If we already have the file and it matches, then just return the target path.
	if verify, _ := d.VerifyChecksum(d.config.TargetPath); verify {
		log.Println("[DEBUG] Initial checksum matched, no download needed.")
		return d.config.TargetPath, nil
	}

	u, err := url.Parse(d.config.Url)
	if err != nil {
		return "", err
	}

	log.Printf("Parsed URL: %#v", u)

	/* FIXME:
		handle the special case of d.config.CopyFile which returns the path
		in an os-specific format.
	*/

	// Files when we don't copy the file are special cased.
	var f *os.File
	var finalPath string
	sourcePath := ""
	if u.Scheme == "file" && !d.config.CopyFile {
		// This is special case for relative path in this case user specify
		// file:../ and after parse destination goes to Opaque
		if u.Path != "" {
			// If url.Path is set just use this
			finalPath = u.Path
		} else if u.Opaque != "" {
			// otherwise try url.Opaque
			finalPath = u.Opaque
		}
		// This is a special case where we use a source file that already exists
		// locally and we don't make a copy. Normally we would copy or download.
		log.Printf("[DEBUG] Using local file: %s", finalPath)

		// transform the actual file uri to a windowsy path if we're being windowsy.
		if runtime.GOOS == "windows" {
			// FIXME: cwd should point to a path relative to the TEMPLATE path,
			//        but since this isn't exposed to us anywhere, we use os.Getwd()
			//        and assume the user ran packer in the same directory that
			//        any relative files are located at.
			cwd,err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("Unable to get working directory")
			}
			finalPath = NormalizeWindowsURL(cwd, *url)
		}

		// Keep track of the source so we can make sure not to delete this later
		sourcePath = finalPath
		if _, err = os.Stat(finalPath); err != nil {
			return "", err
		}
	} else {
		finalPath = d.config.TargetPath

		var ok bool
		d.downloader, ok = d.config.DownloaderMap[u.Scheme]
		if !ok {
			return "", fmt.Errorf("No downloader for scheme: %s", u.Scheme)
		}

		// Otherwise, download using the downloader.
		f, err = os.OpenFile(finalPath, os.O_RDWR|os.O_CREATE, os.FileMode(0666))
		if err != nil {
			return "", err
		}

		log.Printf("[DEBUG] Downloading: %s", u.String())
		err = d.downloader.Download(f, u)
		f.Close()
		if err != nil {
			return "", err
		}
	}

	if d.config.Hash != nil {
		var verify bool
		verify, err = d.VerifyChecksum(finalPath)
		if err == nil && !verify {
			// Only delete the file if we made a copy or downloaded it
			if sourcePath != finalPath {
				os.Remove(finalPath)
			}

			err = fmt.Errorf(
				"checksums didn't match expected: %s",
				hex.EncodeToString(d.config.Checksum))
		}
	}

	return finalPath, err
}

// PercentProgress returns the download progress as a percentage.
func (d *DownloadClient) PercentProgress() int {
	if d.downloader == nil {
		return -1
	}

	return int((float64(d.downloader.Progress()) / float64(d.downloader.Total())) * 100)
}

// VerifyChecksum tests that the path matches the checksum for the
// download.
func (d *DownloadClient) VerifyChecksum(path string) (bool, error) {
	if d.config.Checksum == nil || d.config.Hash == nil {
		return false, errors.New("Checksum or Hash isn't set on download.")
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	log.Printf("Verifying checksum of %s", path)
	d.config.Hash.Reset()
	io.Copy(d.config.Hash, f)
	return bytes.Equal(d.config.Hash.Sum(nil), d.config.Checksum), nil
}

// HTTPDownloader is an implementation of Downloader that downloads
// files over HTTP.
type HTTPDownloader struct {
	progress  uint
	total     uint
	userAgent string
}

func (*HTTPDownloader) Cancel() {
	// TODO(mitchellh): Implement
}

func (d *HTTPDownloader) Download(dst *os.File, src *url.URL) error {
	log.Printf("Starting download over HTTP: %s", src.String())

	// Seek to the beginning by default
	if _, err := dst.Seek(0, 0); err != nil {
		return err
	}

	// Reset our progress
	d.progress = 0

	// Make the request. We first make a HEAD request so we can check
	// if the server supports range queries. If the server/URL doesn't
	// support HEAD requests, we just fall back to GET.
	req, err := http.NewRequest("HEAD", src.String(), nil)
	if err != nil {
		return err
	}

	if d.userAgent != "" {
		req.Header.Set("User-Agent", d.userAgent)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	resp, err := httpClient.Do(req)
	if err == nil && (resp.StatusCode >= 200 && resp.StatusCode < 300) {
		// If the HEAD request succeeded, then attempt to set the range
		// query if we can.
		if resp.Header.Get("Accept-Ranges") == "bytes" {
			if fi, err := dst.Stat(); err == nil {
				if _, err = dst.Seek(0, os.SEEK_END); err == nil {
					req.Header.Set("Range", fmt.Sprintf("bytes=%d-", fi.Size()))
					d.progress = uint(fi.Size())
				}
			}
		}
	}

	// Set the request to GET now, and redo the query to download
	req.Method = "GET"

	resp, err = httpClient.Do(req)
	if err != nil {
		return err
	}

	d.total = d.progress + uint(resp.ContentLength)
	var buffer [4096]byte
	for {
		n, err := resp.Body.Read(buffer[:])
		if err != nil && err != io.EOF {
			return err
		}

		d.progress += uint(n)

		if _, werr := dst.Write(buffer[:n]); werr != nil {
			return werr
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

func (d *HTTPDownloader) Progress() uint {
	return d.progress
}

func (d *HTTPDownloader) Total() uint {
	return d.total
}

// FTPDownloader is an implementation of Downloader that downloads
// files over FTP.
type FTPDownloader struct {
	userInfo url.UserInfo
	mtu uint

	active bool
	progress uint
	total uint
}

func (*FTPDownloader) Cancel() {
	d.active = false
}

func (d *FTPDownloader) Download(dst *os.File, src *url.URL) error {
	var userinfo *url.Userinfo

	userinfo = d.userInfo
	d.active = false

	// check the uri is correct
	uri, err := url.Parse(src)
	if err != nil { return err }

	if uri.Scheme != "ftp" {
		return fmt.Errorf("Unexpected uri scheme: %s", uri.Scheme)
	}

	// connect to ftp server
	var cli *ftp.ServerConn

	log.Printf("Starting download over FTP: %s : %s\n", uri.Host, Uri.Path)
	cli,err := ftp.Dial(uri.Host)
	if err != nil { return nil }
	defer cli.Close()

	// handle authentication
	if uri.User != nil { userinfo = uri.User }

	log.Printf("Authenticating to FTP server: %s : %s\n", uri.User.username, uri.User.password)
	err = cli.Login(userinfo.username, userinfo.password)
	if err != nil { return err }

	// locate specified path
	path := path.Dir(uri.Path)

	log.Printf("Changing to FTP directory : %s\n", path)
	err = cli.ChangeDir(path)
	if err != nil { return nil }

	curpath,err := cli.CurrentDir()
	if err != nil { return err }
	log.Printf("Current FTP directory : %s\n", curpath)

	// collect stats about the specified file
	var name string
	var entry *ftp.Entry

	_,name = path.Split(uri.Path)
	entry = nil

	entries,err := cli.List(curpath)
	for _,e := range entries {
		if e.Type == ftp.EntryTypeFile && e.Name == name {
			entry = e
			break
		}
	}

	if entry == nil {
		return fmt.Errorf("Unable to find file: %s", uri.Path)
	}
	log.Printf("Found file : %s : %v bytes\n", entry.Name, entry.Size)

	d.progress = 0
	d.total = entry.Size

	// download specified file
	d.active = true
	reader,err := cli.RetrFrom(uri.Path, d.progress)
	if err != nil { return nil }

	// do it in a goro so that if someone wants to cancel it, they can
	errch := make(chan error)
	go func(d *FTPDownloader, r *io.Reader, w *bufio.Writer, e chan error) {
		defer w.Flush()
		for ; d.active {
			n,err := io.CopyN(writer, reader, d.mtu)
			if err != nil { break }
			d.progress += n
		}
		d.active = false
		e <- err
	}(d, reader, bufio.NewWriter(dst), errch)

	// spin until it's done
	err = <-errch
	reader.Close()

	if err == nil && d.progress != d.total {
		err = fmt.Errorf("FTP total transfer size was %d when %d was expected", d.progress, d.total)
	}

	// log out and quit
	cli.Logout()
	cli.Quit()
	return err
}

func (d *FTPDownloader) Progress() uint {
	return d.progress
}

func (d *FTPDownloader) Total() uint {
	return d.total
}

// FileDownloader is an implementation of Downloader that downloads
// files using the regular filesystem.
type FileDownloader struct {
	bufferSize *uint

	active bool
	progress uint
	total uint
}

func (*FileDownloader) Cancel() {
	d.active = false
}

func (d *FileDownloader) Progress() uint {
	return d.progress
}

func (d *FileDownloader) Download(dst *os.File, src *url.URL) error {
	d.active = false

	/* parse the uri using the net/url module */
	uri, err := url.Parse(src)
	if uri.Scheme != "file" {
		return fmt.Errorf("Unexpected uri scheme: %s", uri.Scheme)
	}

	/* use the current working directory as the base for relative uri's */
	cwd,err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Unable to get working directory")
	}

	/* determine which uri format is being used and convert to a real path */
	var realpath string, basepath string
	basepath = filepath.ToSlash(cwd)

	// absolute path -- file://c:/absolute/path
	if strings.HasSuffix(uri.Host, ":") {
		realpath = path.Join(uri.Host, uri.Path)

	// semi-absolute path (current drive letter) -- file:///absolute/path
	} else if uri.Host == "" && strings.HasPrefix(uri.Path, "/") {
		realpath = path.Join(filepath.VolumeName(basepath), uri.Path)

	// relative path -- file://./relative/path
	} else if uri.Host == "." {
		realpath = path.Join(basepath, uri.Path)

	// relative path -- file://relative/path
	} else {
		realpath = path.Join(basepath, uri.Host, uri.Path)
	}

	/* download the file using the operating system's facilities */
	d.progress = 0
	d.active = true

	f, err = os.Open(realpath)
	if err != nil { return err }
	defer f.Close()

	// get the file size
	fi, err := f.Stat()
	if err != nil { return err }
	d.total = fi.Size()

	// no bufferSize specified, so copy synchronously.
	if d.bufferSize == nil {
		n,err := io.Copy(dst, f)
		d.active = false
		d.progress += n

	// use a goro in case someone else wants to enable cancel/resume
	} else {
		errch := make(chan error)
		go func(d* FileDownloader, r *bufio.Reader, w *bufio.Writer, e chan error) {
			defer w.Flush()
			for ; d.active {
				n,err := io.CopyN(writer, reader, d.bufferSize)
				if err != nil { break }
				d.progress += n
			}
			d.active = false
			e <- err
		}(d, f, bufio.NewWriter(dst), errch)

		// ...and we spin until it's done
		err = <-errch
	}
	f.Close()
	return err
}

func (d *FileDownloader) Total() uint {
	return d.total
}

// SMBDownloader is an implementation of Downloader that downloads
// files using the "\\" path format on Windows
type SMBDownloader struct {
	bufferSize *uint

	active bool
	progress uint
	total uint
}

func (*SMBDownloader) Cancel() {
	d.active = false
}

func (d *SMBDownloader) Progress() uint {
	return d.progress
}

func (d *SMBDownloader) Download(dst *os.File, src *url.URL) error {
	const UNCPrefix = string(os.PathSeparator)+string(os.PathSeparator)
	d.active = false

	if runtime.GOOS != "windows" {
		return fmt.Errorf("Support for SMB based uri's are not supported on %s", runtime.GOOS)
	}

	/* convert the uri using the net/url module to a UNC path */
	var realpath string
	uri, err := url.Parse(src)
	if uri.Scheme != "smb" {
		return fmt.Errorf("Unexpected uri scheme: %s", uri.Scheme)
	}

	realpath = UNCPrefix + filepath.ToSlash(path.Join(uri.Host, uri.Path))

	/* Open up the "\\"-prefixed path using the Windows filesystem */
	d.progress = 0
	d.active = true

	f, err = os.Open(realpath)
	if err != nil { return err }
	defer f.Close()

	// get the file size (at the risk of performance)
	fi, err := f.Stat()
	if err != nil { return err }
	d.total = fi.Size()

	// no bufferSize specified, so copy synchronously.
	if d.bufferSize == nil {
		n,err := io.Copy(dst, f)
		d.active = false
		d.progress += n

	// use a goro in case someone else wants to enable cancel/resume
	} else {
		errch := make(chan error)
		go func(d* SMBDownloader, r *bufio.Reader, w *bufio.Writer, e chan error) {
			defer w.Flush()
			for ; d.active {
				n,err := io.CopyN(writer, reader, d.bufferSize)
				if err != nil { break }
				d.progress += n
			}
			d.active = false
			e <- err
		}(d, f, bufio.NewWriter(dst), errch)

		// ...and as usual we spin until it's done
		err = <-errch
	}
	f.Close()
	return err
}

func (d *SMBDownloader) Total() uint {
	return d.total
}
