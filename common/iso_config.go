package common

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
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

	if c.RawSingleISOUrl == "" && len(c.ISOUrls) == 0 {
		errs = append(
			errs, errors.New("One of iso_url or iso_urls must be specified."))
	} else if c.RawSingleISOUrl != "" && len(c.ISOUrls) > 0 {
		errs = append(
			errs, errors.New("Only one of iso_url or iso_urls may be specified."))
	} else if c.RawSingleISOUrl != "" {
		c.ISOUrls = []string{c.RawSingleISOUrl}
	}

	if c.ISOChecksumType == "" {
		errs = append(
			errs, errors.New("The iso_checksum_type must be specified."))
	} else {
		c.ISOChecksumType = strings.ToLower(c.ISOChecksumType)
		if c.ISOChecksumType != "none" {
			if c.ISOChecksum == "" {
				errs = append(
					errs, errors.New("Due to large file sizes, an iso_checksum is required"))
				return warnings, errs
			} else {
				if h := HashForType(c.ISOChecksumType); h == nil {
					errs = append(
						errs, fmt.Errorf("Unsupported checksum type: %s", c.ISOChecksumType))
					return warnings, errs
				}

				u, err := url.Parse(c.ISOChecksum)
				if err != nil {
					errs = append(errs,
						fmt.Errorf("Error parsing checksum: %s", err))
					return warnings, errs
				}
				switch u.Scheme {
				case "http", "https", "ftp", "ftps":
					res, err := http.Get(c.ISOChecksum)
					c.ISOChecksum = ""
					if err != nil {
						errs = append(errs,
							fmt.Errorf("Error getting checksum from url: %s", c.ISOChecksum))
						return warnings, errs
					}
					defer res.Body.Close()

					rd := bufio.NewReader(res.Body)
					for {
						line, err := rd.ReadString('\n')
						if err != nil {
							errs = append(errs,
								fmt.Errorf("Error getting checksum from url: %s , %s", c.ISOChecksum, err.Error()))
							return warnings, errs
						}
						parts := strings.Fields(line)
						if len(parts) < 2 {
							continue
						}
						if strings.ToLower(parts[0]) == c.ISOChecksumType {
							// BSD-style checksum
							if parts[1] == fmt.Sprintf("(%s)", filepath.Base(c.ISOUrls[0])) {
								c.ISOChecksum = parts[3]
								break
							}
						} else {
							// Standard checksum
							if parts[1][0] == '*' {
								// Binary mode
								parts[1] = parts[1][1:]
							}
							if parts[1] == filepath.Base(c.ISOUrls[0]) {
								c.ISOChecksum = parts[0]
								break
							}
						}
					}
				case "":
					break
				default:
					errs = append(errs,
						fmt.Errorf("Error parsing checksum url, scheme not supported: %s", u.Scheme))
					return warnings, errs
				}
			}
		}
	}

	c.ISOChecksum = strings.ToLower(c.ISOChecksum)

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
