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
	"log"
	"net/url"
	"os"
	"runtime"
	"path"
	"strings"
)

// imports related to each Downloader implementation
import (
	"io"
	"path/filepath"
	"net/http"
	"github.com/jlaffaye/ftp"
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
			"ftp":   &FTPDownloader{userInfo: url.UserPassword("anonymous", "anonymous@"), mtu: mtu},
			"http":  &HTTPDownloader{userAgent: c.UserAgent},
			"https": &HTTPDownloader{userAgent: c.UserAgent},
			"smb":   &SMBDownloader{bufferSize: nil},
		}
	}

	return &DownloadClient{config: c}
}

// A downloader implements the ability to transfer a file, and cancel or resume
//	it.
type Downloader interface {
	Resume()
	Cancel()
	Progress() uint64
	Total() uint64
}

// A LocalDownloader is responsible for converting a uri to a local path
//	that the platform can open directly.
type LocalDownloader interface {
	toPath(string, url.URL) (string,error)
}

// A RemoteDownloader is responsible for actually taking a remote URL and
//	downloading it.
type RemoteDownloader interface {
	Download(*os.File, *url.URL) error
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

	/* parse the configuration url into a net/url object */
	u, err := url.Parse(d.config.Url)
	if err != nil { return "", err }
	log.Printf("Parsed URL: %#v", u)

	/* use the current working directory as the base for relative uri's */
	cwd,err := os.Getwd()
	if err != nil { return "", err }

	// Determine which is the correct downloader to use
	var finalPath string

	var ok bool
	d.downloader, ok = d.config.DownloaderMap[u.Scheme]
	if !ok {
		return "", fmt.Errorf("No downloader for scheme: %s", u.Scheme)
	}

	remote,ok := d.downloader.(RemoteDownloader)
	if !ok {
		return "", fmt.Errorf("Unable to treat uri scheme %s as a Downloader : %T", u.Scheme, d.downloader)
	}

	local,ok := d.downloader.(LocalDownloader)
	if !ok && !d.config.CopyFile{
		return "", fmt.Errorf("Not allowed to use uri scheme %s in no copy file mode : %T", u.Scheme, d.downloader)
	}

	// If we're copying the file, then just use the actual downloader
	if d.config.CopyFile {
		var f *os.File
		finalPath = d.config.TargetPath

		f, err = os.OpenFile(finalPath, os.O_RDWR|os.O_CREATE, os.FileMode(0666))
		if err != nil { return "", err }

		log.Printf("[DEBUG] Downloading: %s", u.String())
		err = remote.Download(f, u)
		f.Close()
		if err != nil { return "", err }

	// Otherwise if our Downloader is a LocalDownloader we can just use the
	//	path after transforming it.
	} else {
		finalPath,err = local.toPath(cwd, *u)
		if err != nil { return "", err }

		log.Printf("[DEBUG] Using local file: %s", finalPath)
	}

	if d.config.Hash != nil {
		var verify bool
		verify, err = d.VerifyChecksum(finalPath)
		if err == nil && !verify {
			// Only delete the file if we made a copy or downloaded it
			if d.config.CopyFile { os.Remove(finalPath) }

			err = fmt.Errorf(
				"checksums didn't match expected: %s",
				hex.EncodeToString(d.config.Checksum))
		}
	}

	return finalPath, err
}

// PercentProgress returns the download progress as a percentage.
func (d *DownloadClient) PercentProgress() int {
	if d.downloader == nil { return -1 }
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
	progress  uint64
	total     uint64
	userAgent string
}

func (d *HTTPDownloader) Cancel() {
	// TODO(mitchellh): Implement
}

func (d *HTTPDownloader) Resume() {
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
					d.progress = uint64(fi.Size())
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

	d.total = d.progress + uint64(resp.ContentLength)
	var buffer [4096]byte
	for {
		n, err := resp.Body.Read(buffer[:])
		if err != nil && err != io.EOF {
			return err
		}

		d.progress += uint64(n)

		if _, werr := dst.Write(buffer[:n]); werr != nil {
			return werr
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

func (d *HTTPDownloader) Progress() uint64 {
	return d.progress
}

func (d *HTTPDownloader) Total() uint64 {
	return d.total
}

// FTPDownloader is an implementation of Downloader that downloads
// files over FTP.
type FTPDownloader struct {
	userInfo *url.Userinfo
	mtu uint

	active bool
	progress uint64
	total uint64
}

func (d *FTPDownloader) Progress() uint64 {
	return d.progress
}

func (d *FTPDownloader) Total() uint64 {
	return d.total
}

func (d *FTPDownloader) Cancel() {
	d.active = false
}

func (d *FTPDownloader) Resume() {
	// TODO: Implement
}

func (d *FTPDownloader) Download(dst *os.File, src *url.URL) error {
	var userinfo *url.Userinfo

	userinfo = d.userInfo
	d.active = false

	// check the uri is correct
	if src == nil || src.Scheme != "ftp" {
		return fmt.Errorf("Unexpected uri scheme: %s", src.Scheme)
	}
	uri := src

	// connect to ftp server
	var cli *ftp.ServerConn

	log.Printf("Starting download over FTP: %s : %s\n", uri.Host, uri.Path)
	cli,err := ftp.Dial(uri.Host)
	if err != nil { return nil }
	defer cli.Quit()

	// handle authentication
	if uri.User != nil { userinfo = uri.User }

	pass,ok := userinfo.Password()
	if !ok { pass = "ftp@" }

	log.Printf("Authenticating to FTP server: %s : %s\n", userinfo.Username(), pass)
	err = cli.Login(userinfo.Username(), pass)
	if err != nil { return err }

	// locate specified path
	p := path.Dir(uri.Path)

	log.Printf("Changing to FTP directory : %s\n", p)
	err = cli.ChangeDir(p)
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
		if e.Type == ftp.EntryTypeFile && e.Name == name {
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
	go func(d *FTPDownloader, r io.Reader, w io.Writer, e chan error) {
		for ; d.active;  {
			n,err := io.CopyN(w, r, int64(d.mtu))
			if err != nil { break }
			d.progress += uint64(n)
		}
		d.active = false
		e <- err
	}(d, reader, dst, errch)

	// spin until it's done
	err = <-errch
	reader.Close()

	if err == nil && d.progress != d.total {
		err = fmt.Errorf("FTP total transfer size was %d when %d was expected", d.progress, d.total)
	}

	// log out
	cli.Logout()
	return err
}

// FileDownloader is an implementation of Downloader that downloads
// files using the regular filesystem.
type FileDownloader struct {
	bufferSize *uint

	active bool
	progress uint64
	total uint64
}

func (d *FileDownloader) Progress() uint64 {
	return d.progress
}

func (d *FileDownloader) Total() uint64 {
	return d.total
}

func (d *FileDownloader) Cancel() {
	d.active = false
}

func (d *FileDownloader) Resume() {
	// TODO: Implement
}

func (d *FileDownloader) toPath(base string, uri url.URL) (string,error) {
	var result string

	// absolute path -- file://c:/absolute/path -> c:/absolute/path
	if strings.HasSuffix(uri.Host, ":") {
		result = path.Join(uri.Host, uri.Path)

	// semi-absolute path (current drive letter)
	//	-- file:///absolute/path -> /absolute/path
	} else if uri.Host == "" && strings.HasPrefix(uri.Path, "/") {
		result = path.Join(filepath.VolumeName(base), uri.Path)

	// relative path -- file://./relative/path -> ./relative/path
	} else if uri.Host == "." {
		result = path.Join(base, uri.Path)

	// relative path -- file://relative/path -> ./relative/path
	} else {
		result = path.Join(base, uri.Host, uri.Path)
	}
	return filepath.ToSlash(result),nil
}

func (d *FileDownloader) Download(dst *os.File, src *url.URL) error {
	d.active = false

	/* check the uri's scheme to make sure it matches */
	if src == nil || src.Scheme != "file" {
		return fmt.Errorf("Unexpected uri scheme: %s", src.Scheme)
	}
	uri := src

	/* use the current working directory as the base for relative uri's */
	cwd,err := os.Getwd()
	if err != nil { return err }

	/* determine which uri format is being used and convert to a real path */
	realpath,err := d.toPath(cwd, *uri)
	if err != nil { return err }

	/* download the file using the operating system's facilities */
	d.progress = 0
	d.active = true

	f, err := os.Open(realpath)
	if err != nil { return err }
	defer f.Close()

	// get the file size
	fi, err := f.Stat()
	if err != nil { return err }
	d.total = uint64(fi.Size())

	// no bufferSize specified, so copy synchronously.
	if d.bufferSize == nil {
		var n int64
		n,err = io.Copy(dst, f)
		d.active = false
		d.progress += uint64(n)

	// use a goro in case someone else wants to enable cancel/resume
	} else {
		errch := make(chan error)
		go func(d* FileDownloader, r io.Reader, w io.Writer, e chan error) {
			for ; d.active; {
				n,err := io.CopyN(w, r, int64(*d.bufferSize))
				if err != nil { break }
				d.progress += uint64(n)
			}
			d.active = false
			e <- err
		}(d, f, dst, errch)

		// ...and we spin until it's done
		err = <-errch
	}
	f.Close()
	return err
}

// SMBDownloader is an implementation of Downloader that downloads
// files using the "\\" path format on Windows
type SMBDownloader struct {
	bufferSize *uint

	active bool
	progress uint64
	total uint64
}

func (d *SMBDownloader) Progress() uint64 {
	return d.progress
}

func (d *SMBDownloader) Total() uint64 {
	return d.total
}

func (d *SMBDownloader) Cancel() {
	d.active = false
}

func (d *SMBDownloader) Resume() {
	// TODO: Implement
}

func (d *SMBDownloader) toPath(base string, uri url.URL) (string,error) {
	const UNCPrefix = string(os.PathSeparator)+string(os.PathSeparator)

	if runtime.GOOS != "windows" {
		return "",fmt.Errorf("Support for SMB based uri's are not supported on %s", runtime.GOOS)
	}

	return UNCPrefix + filepath.ToSlash(path.Join(uri.Host, uri.Path)), nil
}

func (d *SMBDownloader) Download(dst *os.File, src *url.URL) error {
	d.active = false

	/* convert the uri using the net/url module to a UNC path */
	if src == nil || src.Scheme != "smb" {
		return fmt.Errorf("Unexpected uri scheme: %s", src.Scheme)
	}
	uri := src

	/* use the current working directory as the base for relative uri's */
	cwd,err := os.Getwd()
	if err != nil { return err }

	/* convert uri to an smb-path */
	realpath,err := d.toPath(cwd, *uri)
	if err != nil { return err }

	/* Open up the "\\"-prefixed path using the Windows filesystem */
	d.progress = 0
	d.active = true

	f, err := os.Open(realpath)
	if err != nil { return err }
	defer f.Close()

	// get the file size (at the risk of performance)
	fi, err := f.Stat()
	if err != nil { return err }
	d.total = uint64(fi.Size())

	// no bufferSize specified, so copy synchronously.
	if d.bufferSize == nil {
		var n int64
		n,err = io.Copy(dst, f)
		d.active = false
		d.progress += uint64(n)

	// use a goro in case someone else wants to enable cancel/resume
	} else {
		errch := make(chan error)
		go func(d* SMBDownloader, r io.Reader, w io.Writer, e chan error) {
			for ; d.active; {
				n,err := io.CopyN(w, r, int64(*d.bufferSize))
				if err != nil { break }
				d.progress += uint64(n)
			}
			d.active = false
			e <- err
		}(d, f, dst, errch)

		// ...and as usual we spin until it's done
		err = <-errch
	}
	f.Close()
	return err
}
