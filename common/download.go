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
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/packer/packer" // imports related to each Downloader implementation
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
func NewDownloadClient(c *DownloadConfig, bar packer.ProgressBar) *DownloadClient {
	const mtu = 1500 /* ethernet */ - 20 /* ipv4 */ - 20 /* tcp */

	// Create downloader map if it hasn't been specified already.
	if c.DownloaderMap == nil {
		c.DownloaderMap = map[string]Downloader{
			"file":  &FileDownloader{progress: bar, bufferSize: nil},
			"http":  &HTTPDownloader{progress: bar, userAgent: c.UserAgent},
			"https": &HTTPDownloader{progress: bar, userAgent: c.UserAgent},
			"smb":   &SMBDownloader{progress: bar, bufferSize: nil},
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
	toPath(string, url.URL) (string, error)
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
	if err != nil {
		return "", err
	}
	log.Printf("Parsed URL: %#v", u)

	/* use the current working directory as the base for relative uri's */
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Determine which is the correct downloader to use
	var finalPath string

	var ok bool
	d.downloader, ok = d.config.DownloaderMap[u.Scheme]
	if !ok {
		return "", fmt.Errorf("No downloader for scheme: %s", u.Scheme)
	}

	remote, ok := d.downloader.(RemoteDownloader)
	if !ok {
		return "", fmt.Errorf("Unable to treat uri scheme %s as a Downloader. : %T", u.Scheme, d.downloader)
	}

	local, ok := d.downloader.(LocalDownloader)
	if !ok && !d.config.CopyFile {
		d.config.CopyFile = true
	}

	// If we're copying the file, then just use the actual downloader
	if d.config.CopyFile {
		var f *os.File
		finalPath = d.config.TargetPath

		f, err = os.OpenFile(finalPath, os.O_RDWR|os.O_CREATE, os.FileMode(0666))
		if err != nil {
			return "", err
		}

		log.Printf("[DEBUG] Downloading: %s", u.String())
		err = remote.Download(f, u)
		f.Close()
		if err != nil {
			return "", err
		}

		// Otherwise if our Downloader is a LocalDownloader we can just use the
		//	path after transforming it.
	} else {
		finalPath, err = local.toPath(cwd, *u)
		if err != nil {
			return "", err
		}

		log.Printf("[DEBUG] Using local file: %s", finalPath)
	}

	if d.config.Hash != nil {
		var verify bool
		verify, err = d.VerifyChecksum(finalPath)
		if err == nil && !verify {
			// Only delete the file if we made a copy or downloaded it
			if d.config.CopyFile {
				os.Remove(finalPath)
			}

			err = fmt.Errorf(
				"checksums didn't match expected: %s",
				hex.EncodeToString(d.config.Checksum))
		}
	}

	return finalPath, err
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
	current   uint64
	total     uint64
	userAgent string

	progress packer.ProgressBar
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
	d.current = 0

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
	if err != nil || resp == nil {

		if resp == nil {
			log.Printf("[DEBUG] (download) HTTP connection error: %s", err.Error())

		} else if resp.StatusCode >= 400 && resp.StatusCode < 600 {
			log.Printf("[DEBUG] (download) Non-successful HTTP status code (%s) while making HEAD request: %s", resp.Status, err.Error())

		} else {
			log.Printf("[DEBUG] (download) Error making HTTP HEAD request (%s): %s", resp.Status, err.Error())
		}

	} else {

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// If the HEAD request succeeded, then attempt to set the range
			// query if we can.

			if resp.Header.Get("Accept-Ranges") == "bytes" {
				if fi, err := dst.Stat(); err == nil {
					if _, err = dst.Seek(0, os.SEEK_END); err == nil {
						req.Header.Set("Range", fmt.Sprintf("bytes=%d-", fi.Size()))

						d.current = uint64(fi.Size())
					}
				}
			}

		} else {
			log.Printf("[DEBUG] (download) Unexpected HTTP response during HEAD request: %s", resp.Status)
		}
	}

	// Set the request to GET now, and redo the query to download
	req.Method = "GET"

	resp, err = httpClient.Do(req)
	if err == nil && (resp.StatusCode >= 400 && resp.StatusCode < 600) {
		return fmt.Errorf("Error making HTTP GET request: %s", resp.Status)

	} else if err != nil {
		if resp == nil {
			return fmt.Errorf("HTTP connection error: %s", err.Error())

		} else if resp.StatusCode >= 400 && resp.StatusCode < 600 {
			return fmt.Errorf("HTTP %s error: %s", resp.Status, err.Error())
		}
		return fmt.Errorf("HTTP error: %s", err.Error())
	}

	d.total = d.current + uint64(resp.ContentLength)

	bar := d.progress
	log.Printf("this %#v", bar)
	log.Printf("that")
	bar.Start(d.total)
	bar.Set(d.current)

	var buffer [4096]byte
	for {
		n, err := resp.Body.Read(buffer[:])
		if err != nil && err != io.EOF {
			return err
		}

		d.current += uint64(n)
		bar.Set(d.current)

		if _, werr := dst.Write(buffer[:n]); werr != nil {
			return werr
		}

		if err == io.EOF {
			break
		}
	}
	bar.Finish()
	return nil
}

func (d *HTTPDownloader) Progress() uint64 {
	return d.current
}

func (d *HTTPDownloader) Total() uint64 {
	return d.total
}

// FileDownloader is an implementation of Downloader that downloads
// files using the regular filesystem.
type FileDownloader struct {
	bufferSize *uint

	active  bool
	current uint64
	total   uint64

	progress packer.ProgressBar
}

func (d *FileDownloader) Progress() uint64 {
	return d.current
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

func (d *FileDownloader) toPath(base string, uri url.URL) (string, error) {
	var result string

	// absolute path -- file://c:/absolute/path -> c:/absolute/path
	if strings.HasSuffix(uri.Host, ":") {
		result = path.Join(uri.Host, uri.Path)

		// semi-absolute path (current drive letter)
		//	-- file:///absolute/path -> drive:/absolute/path
	} else if uri.Host == "" && strings.HasPrefix(uri.Path, "/") {
		apath := uri.Path
		components := strings.Split(apath, "/")
		volume := filepath.VolumeName(base)

		// semi-absolute absolute path (includes volume letter)
		// -- file://drive:/path -> drive:/absolute/path
		if len(components) > 1 && strings.HasSuffix(components[1], ":") {
			volume = components[1]
			apath = path.Join(components[2:]...)
		}

		result = path.Join(volume, apath)

		// relative path -- file://./relative/path -> ./relative/path
	} else if uri.Host == "." {
		result = path.Join(base, uri.Path)

		// relative path -- file://relative/path -> ./relative/path
	} else {
		result = path.Join(base, uri.Host, uri.Path)
	}
	return filepath.ToSlash(result), nil
}

func (d *FileDownloader) Download(dst *os.File, src *url.URL) error {
	d.active = false

	/* check the uri's scheme to make sure it matches */
	if src == nil || src.Scheme != "file" {
		return fmt.Errorf("Unexpected uri scheme: %s", src.Scheme)
	}
	uri := src

	/* use the current working directory as the base for relative uri's */
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	/* determine which uri format is being used and convert to a real path */
	realpath, err := d.toPath(cwd, *uri)
	if err != nil {
		return err
	}

	/* download the file using the operating system's facilities */
	d.current = 0
	d.active = true

	f, err := os.Open(realpath)
	if err != nil {
		return err
	}
	defer f.Close()

	// get the file size
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	d.total = uint64(fi.Size())

	bar := d.progress

	bar.Start(d.total)
	bar.Set(d.current)

	// no bufferSize specified, so copy synchronously.
	if d.bufferSize == nil {
		var n int64
		n, err = io.Copy(dst, f)
		d.active = false

		d.current += uint64(n)
		bar.Set(d.current)

		// use a goro in case someone else wants to enable cancel/resume
	} else {
		errch := make(chan error)
		go func(d *FileDownloader, r io.Reader, w io.Writer, e chan error) {
			for d.active {
				n, err := io.CopyN(w, r, int64(*d.bufferSize))
				if err != nil {
					break
				}

				d.current += uint64(n)
				bar.Set(d.current)
			}
			d.active = false
			e <- err
		}(d, f, dst, errch)

		// ...and we spin until it's done
		err = <-errch
	}
	bar.Finish()
	f.Close()
	return err
}

// SMBDownloader is an implementation of Downloader that downloads
// files using the "\\" path format on Windows
type SMBDownloader struct {
	bufferSize *uint

	active  bool
	current uint64
	total   uint64

	progress packer.ProgressBar
}

func (d *SMBDownloader) Progress() uint64 {
	return d.current
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

func (d *SMBDownloader) toPath(base string, uri url.URL) (string, error) {
	const UNCPrefix = string(os.PathSeparator) + string(os.PathSeparator)

	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("Support for SMB based uri's are not supported on %s", runtime.GOOS)
	}

	return UNCPrefix + filepath.ToSlash(path.Join(uri.Host, uri.Path)), nil
}

func (d *SMBDownloader) Download(dst *os.File, src *url.URL) error {

	/* first we warn the world if we're not running windows */
	if runtime.GOOS != "windows" {
		return fmt.Errorf("Support for SMB based uri's are not supported on %s", runtime.GOOS)
	}

	d.active = false

	/* convert the uri using the net/url module to a UNC path */
	if src == nil || src.Scheme != "smb" {
		return fmt.Errorf("Unexpected uri scheme: %s", src.Scheme)
	}
	uri := src

	/* use the current working directory as the base for relative uri's */
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	/* convert uri to an smb-path */
	realpath, err := d.toPath(cwd, *uri)
	if err != nil {
		return err
	}

	/* Open up the "\\"-prefixed path using the Windows filesystem */
	d.current = 0
	d.active = true

	f, err := os.Open(realpath)
	if err != nil {
		return err
	}
	defer f.Close()

	// get the file size (at the risk of performance)
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	d.total = uint64(fi.Size())

	bar := d.progress

	bar.Start(d.current)

	// no bufferSize specified, so copy synchronously.
	if d.bufferSize == nil {
		var n int64
		n, err = io.Copy(dst, f)
		d.active = false

		d.current += uint64(n)
		bar.Set(d.current)

		// use a goro in case someone else wants to enable cancel/resume
	} else {
		errch := make(chan error)
		go func(d *SMBDownloader, r io.Reader, w io.Writer, e chan error) {
			for d.active {
				n, err := io.CopyN(w, r, int64(*d.bufferSize))
				if err != nil {
					break
				}

				d.current += uint64(n)
				bar.Set(d.current)
			}
			d.active = false
			e <- err
		}(d, f, dst, errch)

		// ...and as usual we spin until it's done
		err = <-errch
	}
	bar.Finish()
	f.Close()
	return err
}
