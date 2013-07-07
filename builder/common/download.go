package common

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
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
			"http": new(HTTPDownloader),
		}
	}

	return &DownloadClient{config: c}
}

// A downloader is responsible for actually taking a remote URL and
// downloading it.
type Downloader interface {
	Cancel()
	Download(io.Writer, *url.URL) error
	Progress() uint
	Total() uint
}

func (d *DownloadClient) Cancel() {
	// TODO(mitchellh): Implement
}

func (d *DownloadClient) Get() (string, error) {
	// If we already have the file and it matches, then just return the target path.
	if verify, _ := d.VerifyChecksum(d.config.TargetPath); verify {
		log.Println("Initial checksum matched, no download needed.")
		return d.config.TargetPath, nil
	}

	url, err := url.Parse(d.config.Url)
	if err != nil {
		return "", err
	}

	log.Printf("Parsed URL: %#v", url)

	// Files when we don't copy the file are special cased.
	var finalPath string
	if url.Scheme == "file" && !d.config.CopyFile {
		finalPath = url.Path
	} else {
		finalPath = d.config.TargetPath

		var ok bool
		d.downloader, ok = d.config.DownloaderMap[url.Scheme]
		if !ok {
			return "", fmt.Errorf("No downloader for scheme: %s", url.Scheme)
		}

		// Otherwise, download using the downloader.
		f, err := os.Create(finalPath)
		if err != nil {
			return "", err
		}
		defer f.Close()

		log.Printf("Downloading: %s", url.String())
		err = d.downloader.Download(f, url)
		if err != nil {
			return "", err
		}
	}

	if d.config.Hash != nil {
		var verify bool
		verify, err = d.VerifyChecksum(finalPath)
		if err == nil && !verify {
			err = fmt.Errorf("checksums didn't match expected: %s", hex.EncodeToString(d.config.Checksum))
		}
	}

	return finalPath, err
}

// PercentProgress returns the download progress as a percentage.
func (d *DownloadClient) PercentProgress() uint {
	if d.downloader == nil {
		return 0
	}

	return uint((float64(d.downloader.Progress()) / float64(d.downloader.Total())) * 100)
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
	return bytes.Compare(d.config.Hash.Sum(nil), d.config.Checksum) == 0, nil
}

// HTTPDownloader is an implementation of Downloader that downloads
// files over HTTP.
type HTTPDownloader struct {
	progress uint
	total    uint
}

func (*HTTPDownloader) Cancel() {
	// TODO(mitchellh): Implement
}

func (d *HTTPDownloader) Download(dst io.Writer, src *url.URL) error {
	log.Printf("Starting download: %s", src.String())
	resp, err := http.Get(src.String())
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		log.Printf(
			"Non-200 status code: %d. Getting error body.", resp.StatusCode)

		errorBody := new(bytes.Buffer)
		io.Copy(errorBody, resp.Body)
		return fmt.Errorf("HTTP error '%d'! Remote side responded:\n%s",
			resp.StatusCode, errorBody.String())
	}

	d.progress = 0
	d.total = uint(resp.ContentLength)

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
