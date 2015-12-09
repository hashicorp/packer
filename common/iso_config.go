package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/packer/template/interpolate"
)

// ISOConfig contains configuration for downloading ISO images.
type ISOConfig struct {
	ISOChecksum     string   `mapstructure:"iso_checksum"`
	ISOChecksumType string   `mapstructure:"iso_checksum_type"`
	ISOUrls         []string `mapstructure:"iso_urls"`
	TargetPath      string   `mapstructure:"iso_target_path"`
	RawSingleISOUrl string   `mapstructure:"iso_url"`
}

func (c *ISOConfig) Prepare(ctx *interpolate.Context) ([]string, []error) {
	// Validation
	var errs []error
	var err error
	var warnings []string

	if c.ISOChecksumType == "" {
		errs = append(
			errs, errors.New("The iso_checksum_type must be specified."))
	} else {
		c.ISOChecksumType = strings.ToLower(c.ISOChecksumType)
		if c.ISOChecksumType != "none" {
			if c.ISOChecksum == "" {
				errs = append(
					errs, errors.New("Due to large file sizes, an iso_checksum is required"))
			} else {
				c.ISOChecksum = strings.ToLower(c.ISOChecksum)
			}

			if h := HashForType(c.ISOChecksumType); h == nil {
				errs = append(
					errs,
					fmt.Errorf("Unsupported checksum type: %s", c.ISOChecksumType))
			}
		}
	}

	if c.RawSingleISOUrl == "" && len(c.ISOUrls) == 0 {
		errs = append(
			errs, errors.New("One of iso_url or iso_urls must be specified."))
	} else if c.RawSingleISOUrl != "" && len(c.ISOUrls) > 0 {
		errs = append(
			errs, errors.New("Only one of iso_url or iso_urls may be specified."))
	} else if c.RawSingleISOUrl != "" {
		c.ISOUrls = []string{c.RawSingleISOUrl}
	}

	for i, url := range c.ISOUrls {
		c.ISOUrls[i], err = DownloadableURL(url)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Failed to parse iso_url %d: %s", i+1, err))
		}
	}

	// Warnings
	if c.ISOChecksumType == "none" {
		warnings = append(warnings,
			"A checksum type of 'none' was specified. Since ISO files are so big,\n"+
				"a checksum is highly recommended.")
	}

	return warnings, errs
}
