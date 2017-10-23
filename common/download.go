package common

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
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

// NewDownloadClient returns a new DownloadClient for the given
// configuration.
func NewDownloadClient(c *DownloadConfig) *DownloadClient {
	if c.DownloaderMap == nil {
		c.DownloaderMap = map[string]Downloader{
			"http":  &HTTPDownloader{userAgent: c.UserAgent},
			"https": &HTTPDownloader{userAgent: c.UserAgent},
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

		// Remove forward slash on absolute Windows file URLs before processing
		if runtime.GOOS == "windows" && len(finalPath) > 0 && finalPath[0] == '/' {
			finalPath = finalPath[1:]
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
		// TODO: MEGAN: Add hashstring to end of URL
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
	log.Printf("Starting download: %s", src.String())

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
