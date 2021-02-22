//go:generate struct-markdown

package commonsteps

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	getter "github.com/hashicorp/go-getter/v2"
	urlhelper "github.com/hashicorp/go-getter/v2/helper/url"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// By default, Packer will symlink, download or copy image files to the Packer
// cache into a "`hash($wim_url+$wim_checksum).$wim_target_extension`" file.
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
// go-getter can guess the checksum type based on `wim_checksum` length, and it is
// also possible to specify the checksum type.
//
// In JSON:
//
// ```json
//   "wim_checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2",
//   "wim_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.wim"
// ```
//
// ```json
//   "wim_checksum": "file:ubuntu.org/..../ubuntu-14.04.1-server-amd64.wim.sum",
//   "wim_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.wim"
// ```
//
// ```json
//   "wim_checksum": "file://./shasums.txt",
//   "wim_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.wim"
// ```
//
// ```json
//   "wim_checksum": "file:./shasums.txt",
//   "wim_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.wim"
// ```
//
// In HCL2:
//
// ```hcl
//   wim_checksum = "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2"
//   wim_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.wim"
// ```
//
// ```hcl
//   wim_checksum = "file:ubuntu.org/..../ubuntu-14.04.1-server-amd64.wim.sum"
//   wim_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.wim"
// ```
//
// ```hcl
//   wim_checksum = "file://./shasums.txt"
//   wim_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.wim"
// ```
//
// ```hcl
//   wim_checksum = "file:./shasums.txt",
//   wim_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.wim"
// ```
//
type WIMConfig struct {
	// The checksum for the WIM file or virtual hard drive file. The type of
	// the checksum is specified within the checksum field as a prefix, ex:
	// "md5:{$checksum}". The type of the checksum can also be omitted and
	// Packer will try to infer it based on string length. Valid values are
	// "none", "{$checksum}", "md5:{$checksum}", "sha1:{$checksum}",
	// "sha256:{$checksum}", "sha512:{$checksum}" or "file:{$path}". Here is a
	// list of valid checksum values:
	//  * md5:090992ba9fd140077b0661cb75f7ce13
	//  * 090992ba9fd140077b0661cb75f7ce13
	//  * sha1:ebfb681885ddf1234c18094a45bbeafd91467911
	//  * ebfb681885ddf1234c18094a45bbeafd91467911
	//  * sha256:ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
	//  * ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
	//  * file:http://releases.ubuntu.com/20.04/SHA256SUMS
	//  * file:file://./local/path/file.sum
	//  * file:./local/path/file.sum
	//  * none
	// Although the checksum will not be verified when it is set to "none",
	// this is not recommended since these files can be very large and
	// corruption does happen from time to time.
	WIMChecksum string `mapstructure:"wim_checksum" required:"true"`
	// A URL to the WIM containing the installation image or virtual hard drive
	// (VHD or VHDX) file to clone.
	RawSingleWIMUrl string `mapstructure:"wim_url" required:"true"`
	// Multiple URLs for the WIM to download. Packer will try these in order.
	// If anything goes wrong attempting to download or while downloading a
	// single URL, it will move on to the next. All URLs must point to the same
	// file (same checksum). By default this is empty and `wim_url` is used.
	// Only one of `wim_url` or `wim_urls` can be specified.
	WIMUrls []string `mapstructure:"wim_urls"`
	// The path where the wim should be saved after download. By default will
	// go in the packer cache, with a hash of the original filename and
	// checksum as its name.
	TargetPath string `mapstructure:"wim_target_path"`
	// The extension of the wim file after download. This defaults to `wim`.
	TargetExtension string `mapstructure:"wim_target_extension"`
}

func (c *WIMConfig) Prepare(*interpolate.Context) (warnings []string, errs []error) {
	if len(c.WIMUrls) != 0 && c.RawSingleWIMUrl != "" {
		errs = append(
			errs, errors.New("Only one of wim_url or wim_urls must be specified"))
		return
	}

	if c.RawSingleWIMUrl != "" {
		// make sure only array is set
		c.WIMUrls = append([]string{c.RawSingleWIMUrl}, c.WIMUrls...)
		c.RawSingleWIMUrl = ""
	}

	if len(c.WIMUrls) == 0 {
		errs = append(
			errs, errors.New("One of wim_url or wim_urls must be specified"))
		return
	}
	if c.TargetExtension == "" {
		c.TargetExtension = "iso"
	}
	c.TargetExtension = strings.ToLower(c.TargetExtension)

	// Warnings
	if c.WIMChecksum == "none" {
		warnings = append(warnings,
			"A checksum of 'none' was specified. Since WIM files are so big,\n"+
				"a checksum is highly recommended.")
		return warnings, errs
	} else if c.WIMChecksum == "" {
		errs = append(errs, fmt.Errorf("A checksum must be specified"))
	} else {
		// ESX5Driver.VerifyChecksum is ran remotely but should not download a
		// checksum file, therefore in case it is a file, we need to download
		// it now and compute the checksum now, we transform it back to a
		// checksum string so that it can be simply read in the VerifyChecksum.
		//
		// Doing this also has the added benefit of failing early if a checksum
		// is incorrect or if getting it should fail.
		u, err := urlhelper.Parse(c.WIMUrls[0])
		if err != nil {
			return warnings, append(errs, fmt.Errorf("url parse: %s", err))
		}

		q := u.Query()
		if c.WIMChecksum != "" {
			q.Set("checksum", c.WIMChecksum)
		}
		u.RawQuery = q.Encode()

		wd, err := os.Getwd()
		if err != nil {
			log.Printf("Getwd: %v", err)
			// here we ignore the error in case the
			// working directory is not needed.
		}

		req := &getter.Request{
			Src: u.String(),
			Pwd: wd,
		}
		cksum, err := defaultGetterClient.GetChecksum(context.TODO(), req)
		if err != nil {
			errs = append(errs, fmt.Errorf("%v in %q", err, req.URL().Query().Get("checksum")))
		} else {
			c.WIMChecksum = cksum.String()
		}
	}

	if strings.HasSuffix(strings.ToLower(c.WIMChecksum), ".iso") {
		errs = append(errs, fmt.Errorf("Error parsing checksum:"+
			" .iso is not a valid checksum ending"))
	}

	return warnings, errs
}
