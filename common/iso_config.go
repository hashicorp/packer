//go:generate struct-markdown

package common

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/template/interpolate"
)

// By default, Packer will symlink, download or copy image files to the Packer
// cache into a "`hash($iso_url+$iso_checksum).$iso_target_extension`" file.
// Packer uses [hashicorp/go-getter](https://github.com/hashicorp/go-getter) in
// file mode in order to perform a download.
//
// go-getter supports the following protocols:
//
// * Local files
// * Git
// * Mercurial
// * HTTP
// * Amazon S3
//
// Examples:
// go-getter can guess the checksum type based on `iso_checksum` len.
//
// ```json
// {
//   "iso_checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2",
//   "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
// }
// ```
//
// ```json
// {
//   "iso_checksum_type": "file",
//   "iso_checksum": "ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum",
//   "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
// }
// ```
//
// ```json
// {
//   "iso_checksum_url": "./shasums.txt",
//   "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
// }
// ```
//
// ```json
// {
//   "iso_checksum_type": "sha256",
//   "iso_checksum_url": "./shasums.txt",
//   "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
// }
// ```
//
type ISOConfig struct {
	// The checksum for the ISO file or virtual hard drive file. The algorithm
	// to use when computing the checksum can be optionally specified with
	// `iso_checksum_type`. When `iso_checksum_type` is not set packer will
	// guess the checksumming type based on `iso_checksum` length.
	// `iso_checksum` can be also be a file or an URL, in which case
	// `iso_checksum_type` must be set to `file`; the go-getter will download
	// it and use the first hash found.
	ISOChecksum string `mapstructure:"iso_checksum" required:"true"`
	// An URL to a checksum file containing a checksum for the ISO file. At
	// least one of `iso_checksum` and `iso_checksum_url` must be defined.
	// `iso_checksum_url` will be ignored if `iso_checksum` is non empty.
	ISOChecksumURL string `mapstructure:"iso_checksum_url"`
	// The algorithm to be used when computing the checksum of the file
	// specified in `iso_checksum`. Currently, valid values are "", "none",
	// "md5", "sha1", "sha256", "sha512" or "file". Since the validity of ISO
	// and virtual disk files are typically crucial to a successful build,
	// Packer performs a check of any supplied media by default. While setting
	// "none" will cause Packer to skip this check, corruption of large files
	// such as ISOs and virtual hard drives can occur from time to time. As
	// such, skipping this check is not recommended. `iso_checksum_type` must
	// be set to `file` when `iso_checksum` is an url.
	ISOChecksumType string `mapstructure:"iso_checksum_type"`
	// A URL to the ISO containing the installation image or virtual hard drive
	// (VHD or VHDX) file to clone.
	RawSingleISOUrl string `mapstructure:"iso_url" required:"true"`
	// Multiple URLs for the ISO to download. Packer will try these in order.
	// If anything goes wrong attempting to download or while downloading a
	// single URL, it will move on to the next. All URLs must point to the same
	// file (same checksum). By default this is empty and `iso_url` is used.
	// Only one of `iso_url` or `iso_urls` can be specified.
	ISOUrls []string `mapstructure:"iso_urls"`
	// The path where the iso should be saved after download. By default will
	// go in the packer cache, with a hash of the original filename and
	// checksum as its name.
	TargetPath string `mapstructure:"iso_target_path"`
	// The extension of the iso file after download. This defaults to `iso`.
	TargetExtension string `mapstructure:"iso_target_extension"`
}

func (c *ISOConfig) Prepare(*interpolate.Context) (warnings []string, errs []error) {
	if len(c.ISOUrls) != 0 && c.RawSingleISOUrl != "" {
		errs = append(
			errs, errors.New("Only one of iso_url or iso_urls must be specified"))
		return
	}

	if c.RawSingleISOUrl != "" {
		// make sure only array is set
		c.ISOUrls = append([]string{c.RawSingleISOUrl}, c.ISOUrls...)
		c.RawSingleISOUrl = ""
	}
	if len(c.ISOUrls) == 0 {
		errs = append(
			errs, errors.New("One of iso_url or iso_urls must be specified"))
		return
	}

	c.ISOChecksumType = strings.ToLower(c.ISOChecksumType)

	if c.TargetExtension == "" {
		c.TargetExtension = "iso"
	}
	c.TargetExtension = strings.ToLower(c.TargetExtension)

	// Warnings
	if c.ISOChecksumType == "none" {
		warnings = append(warnings,
			"A checksum type of 'none' was specified. Since ISO files are so big,\n"+
				"a checksum is highly recommended.")
		return warnings, errs
	}

	if c.ISOChecksumURL != "" {
		if c.ISOChecksum != "" {
			warnings = append(warnings, "You have provided both an "+
				"iso_checksum and an iso_checksum_url. Discarding the "+
				"iso_checksum_url and using the checksum.")
		} else {
			if strings.HasSuffix(strings.ToLower(c.ISOChecksumURL), ".iso") {
				errs = append(errs, fmt.Errorf("Error parsing checksum:"+
					" .iso is not a valid checksum extension"))
			}
			// go-getter auto-parses checksum files
			c.ISOChecksumType = "file"
			c.ISOChecksum = c.ISOChecksumURL
		}
	}

	if c.ISOChecksum == "" {
		errs = append(errs, fmt.Errorf("A checksum must be specified"))
	}
	if c.ISOChecksumType == "file" {
		url := c.ISOChecksum
		if c.ISOChecksumURL != "" {
			url = c.ISOChecksumURL
		}
		cksum, err := defaultGetterClient.ChecksumFromFile(context.TODO(), url, c.ISOUrls[0])
		if err != nil {
			errs = append(errs, fmt.Errorf("Couldn't extract checksum from checksum file: %v", err))
		} else {
			c.ISOChecksumType = cksum.Type
			c.ISOChecksum = hex.EncodeToString(cksum.Value)
		}
	}

	return warnings, errs
}
