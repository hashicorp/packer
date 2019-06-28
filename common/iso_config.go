package common

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/hashicorp/packer/template/interpolate"
)

// ISOConfig contains configuration for downloading ISO images.
type ISOConfig struct {
	ISOChecksum     string   `mapstructure:"iso_checksum"`
	ISOChecksumURL  string   `mapstructure:"iso_checksum_url"`
	ISOChecksumType string   `mapstructure:"iso_checksum_type"`
	ISOUrls         []string `mapstructure:"iso_urls"`
	TargetPath      string   `mapstructure:"iso_target_path"`
	TargetExtension string   `mapstructure:"iso_target_extension"`
	RawSingleISOUrl string   `mapstructure:"iso_url"`
}

func (c *ISOConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {
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
		u, err := url.Parse(c.ISOUrls[0])
		wd, err := os.Getwd()
		if err != nil {
			log.Printf("get working directory: %v", err)
			// here we ignore the error in case the
			// working directory is not needed.
		}
		gc := getter.Client{
			Dst:     "no-op",
			Src:     u.String(),
			Pwd:     wd,
			Dir:     false,
			Getters: getter.Getters,
		}
		cksum, err := gc.ChecksumFromFile(c.ISOChecksumURL, u)
		if cksum == nil || err != nil {
			errs = append(errs, fmt.Errorf("Couldn't extract checksum from checksum file"))
		} else {
			c.ISOChecksumType = cksum.Type
			c.ISOChecksum = hex.EncodeToString(cksum.Value)
		}
	}

	return warnings, errs
}
