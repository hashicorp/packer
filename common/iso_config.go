package common

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
	if c.RawSingleISOUrl == "" && len(c.ISOUrls) == 0 {
		errs = append(
			errs, errors.New("One of iso_url or iso_urls must be specified."))
		return
	} else if c.RawSingleISOUrl != "" && len(c.ISOUrls) > 0 {
		errs = append(
			errs, errors.New("Only one of iso_url or iso_urls may be specified."))
		return
	} else if c.RawSingleISOUrl != "" {
		c.ISOUrls = []string{c.RawSingleISOUrl}
	}

	if c.ISOChecksumType == "" {
		errs = append(
			errs, errors.New("The iso_checksum_type must be specified."))
	} else {
		c.ISOChecksumType = strings.ToLower(c.ISOChecksumType)
		if c.ISOChecksumType != "none" {
			if c.ISOChecksum == "" && c.ISOChecksumURL == "" {
				errs = append(
					errs, errors.New("Due to large file sizes, an iso_checksum is required"))
				return warnings, errs
			} else {
				if h := HashForType(c.ISOChecksumType); h == nil {
					errs = append(
						errs, fmt.Errorf("Unsupported checksum type: %s", c.ISOChecksumType))
					return warnings, errs
				}

				// If iso_checksum has no value use iso_checksum_url instead.
				if c.ISOChecksum == "" {
					if strings.HasSuffix(strings.ToLower(c.ISOChecksumURL), ".iso") {
						errs = append(errs, fmt.Errorf("Error parsing checksum:"+
							" .iso is not a valid checksum extension"))
						return warnings, errs
					}
					u, err := url.Parse(c.ISOChecksumURL)
					if err != nil {
						errs = append(errs,
							fmt.Errorf("Error parsing checksum: %s", err))
						return warnings, errs
					}
					switch u.Scheme {
					case "http", "https":
						res, err := http.Get(c.ISOChecksumURL)
						c.ISOChecksum = ""
						if err != nil {
							errs = append(errs,
								fmt.Errorf("Error getting checksum from url: %s", c.ISOChecksumURL))
							return warnings, errs
						}
						defer res.Body.Close()
						err = c.parseCheckSumFile(bufio.NewReader(res.Body))
						if err != nil {
							errs = append(errs, err)
							return warnings, errs
						}
					case "file":
						path := u.Path

						if runtime.GOOS == "windows" && len(path) > 2 && path[0] == '/' && path[2] == ':' {
							path = strings.TrimLeft(path, "/")
						}

						file, err := os.Open(path)
						if err != nil {
							errs = append(errs, err)
							return warnings, errs
						}
						err = c.parseCheckSumFile(bufio.NewReader(file))
						if err != nil {
							errs = append(errs, err)
							return warnings, errs
						}
					case "":
						break
					default:
						errs = append(errs,
							fmt.Errorf("Error parsing checksum url: %s, scheme not supported: %s", c.ISOChecksumURL, u.Scheme))
						return warnings, errs
					}
				}
			}
		}
	}

	c.ISOChecksum = strings.ToLower(c.ISOChecksum)

	for i, url := range c.ISOUrls {
		url, err := ValidatedURL(url)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Failed to parse iso_url %d: %s", i+1, err))
		} else {
			c.ISOUrls[i] = url
		}
	}

	if c.TargetExtension == "" {
		c.TargetExtension = "iso"
	}
	c.TargetExtension = strings.ToLower(c.TargetExtension)

	// Warnings
	if c.ISOChecksumType == "none" {
		warnings = append(warnings,
			"A checksum type of 'none' was specified. Since ISO files are so big,\n"+
				"a checksum is highly recommended.")
	}

	return warnings, errs
}

func (c *ISOConfig) parseCheckSumFile(rd *bufio.Reader) error {
	u, err := url.Parse(c.ISOUrls[0])
	if err != nil {
		return err
	}

	checksumurl, err := url.Parse(c.ISOChecksumURL)
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(u.Path)
	if err != nil {
		log.Printf("Unable to generate absolute path from provided iso_url: %s", err)
		absPath = ""
	}

	relpath, err := filepath.Rel(filepath.Dir(checksumurl.Path), absPath)
	if err != nil {
		log.Printf("Unable to determine relative pathing; continuing with abspath.")
		relpath = ""
	}

	filename := filepath.Base(u.Path)

	errNotFound := fmt.Errorf("No checksum for %q, %q or %q found at: %s",
		filename, relpath, u.Path, c.ISOChecksumURL)
	for {
		line, err := rd.ReadString('\n')
		if err != nil && line == "" {
			break
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		options := []string{filename, relpath, "./" + relpath, absPath}
		if strings.ToLower(parts[0]) == c.ISOChecksumType {
			// BSD-style checksum
			for _, match := range options {
				if parts[1] == fmt.Sprintf("(%s)", match) {
					c.ISOChecksum = parts[3]
					return nil
				}
			}
		} else {
			// Standard checksum
			if parts[1][0] == '*' {
				// Binary mode
				parts[1] = parts[1][1:]
			}
			for _, match := range options {
				if parts[1] == match {
					c.ISOChecksum = parts[0]
					return nil
				}
			}
		}
	}
	return errNotFound
}
