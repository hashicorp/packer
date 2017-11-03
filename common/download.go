package common

import (
	"encoding/hex"
	"fmt"
	"hash"
	"log"
	"os"

	getter "github.com/hashicorp/go-getter"
)

// DownloadConfig is the configuration given to instantiate a new
// download instance. Once a configuration is used to instantiate
// a download client, it must not be modified.
type DownloadConfig struct {
	// The source URL in the form of a string.
	Url string

	// This is the path to download the file to.
	TargetPath string

	// If true, this will copy even a local file to the target
	// location. If false, then it will "download" the file by just
	// returning the local path to the file.
	CopyFile bool

	// The hashing implementation to use to checksum the downloaded file.
	Hash hash.Hash

	// The type of hashing implementation to use; e.g. "md5"
	HashType string

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
	config       *DownloadConfig
	getterClient *getter.Client
}

// NewDownloadClient returns a new DownloadClient for the given
// configuration.
func NewDownloadClient(c *DownloadConfig) *DownloadClient {

	return &DownloadClient{config: c}
}

func (d *DownloadClient) Cancel() {
	// TODO(mitchellh): Implement
}

func (d *DownloadClient) Get() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Check that the fild hasn't already been downloaded
	checksumMatches := getter.CompareChecksum(d.config.TargetPath,
		d.config.Hash, d.config.Checksum)
	if checksumMatches {
		log.Printf("No need to download anew; given checksum matches the " +
			"file currently at dst.")
		return d.config.TargetPath, nil
	}
	d.config.Hash.Reset()

	// Format src string with checksum for go-getter
	srcPlusChecksum := fmt.Sprintf("%s?checksum=%s:%s", d.config.Url,
		d.config.HashType, hex.EncodeToString(d.config.Checksum))

	// Download fi
	d.getterClient = &getter.Client{
		Src:  srcPlusChecksum,
		Dst:  d.config.TargetPath,
		Pwd:  pwd,
		Mode: getter.ClientModeFile,
		Dir:  false}

	err = d.getterClient.Get()
	if err != nil {
		log.Printf("Error Getting URL: %s", err)
		return "", err
	}

	return d.config.TargetPath, err
}

func (d *DownloadClient) GetDownloadProgress() int {
	return d.getterClient.PercentComplete
}
