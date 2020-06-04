package getter

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	urlhelper "github.com/hashicorp/go-getter/v2/helper/url"
	"github.com/hashicorp/go-multierror"
	safetemp "github.com/hashicorp/go-safetemp"
)

// Client is a client for downloading things.
//
// Top-level functions such as Get are shortcuts for interacting with a client.
// Using a client directly allows more fine-grained control over how downloading
// is done, as well as customizing the protocols supported.
type Client struct {
	// Decompressors is the map of decompressors supported by this client.
	// If this is nil, then the default value is the Decompressors global.
	Decompressors map[string]Decompressor

	// Getters is the list of protocols supported by this client. If this
	// is nil, then the default Getters variable will be used.
	Getters []Getter
}

// GetResult is the result of a Client.Get
type GetResult struct {
	// Local destination of the gotten object.
	Dst string
}

// Get downloads the configured source to the destination.
func (c *Client) Get(ctx context.Context, req *Request) (*GetResult, error) {
	if err := c.configure(); err != nil {
		return nil, err
	}

	// Store this locally since there are cases we swap this
	if req.Mode == ModeInvalid {
		req.Mode = ModeAny
	}

	// If there is a subdir component, then we download the root separately
	// and then copy over the proper subdir.
	req.Src, req.subDir = SourceDirSubdir(req.Src)
	if req.subDir != "" {
		td, tdcloser, err := safetemp.Dir("", "getter")
		if err != nil {
			return nil, err
		}
		defer tdcloser.Close()

		req.realDst = req.Dst
		req.Dst = td
	}

	var multierr []error
	for _, g := range c.Getters {
		shouldDownload, err := Detect(req, g)
		if err != nil {
			return nil, err
		}
		if !shouldDownload {
			// the request should not be processed by that getter
			continue
		}

		result, getErr := c.get(ctx, req, g)
		if getErr != nil {
			if getErr.Fatal {
				return nil, getErr.Err
			}
			multierr = append(multierr, getErr.Err)
			continue
		}

		return result, nil
	}

	if len(multierr) == 1 {
		// This is for keeping the error original format
		return nil, multierr[0]
	}

	if multierr != nil {
		var result *multierror.Error
		result = multierror.Append(result, multierr...)
		return nil, fmt.Errorf("error downloading '%s': %s", req.Src, result.Error())
	}

	return nil, fmt.Errorf("error downloading '%s'", req.Src)
}

// getError is the Error response object returned by get(context.Context, *Request, Getter)
// to tell the client whether to halt (Fatal) Get or to keep trying to get an artifact.
type getError struct {
	// When Fatal is true something went wrong with get(context.Context, *Request, Getter)
	// and the client should halt and return the Err.
	Fatal bool
	Err   error
}

func (ge *getError) Error() string {
	return ge.Err.Error()
}

func (c *Client) get(ctx context.Context, req *Request, g Getter) (*GetResult, *getError) {
	u, err := urlhelper.Parse(req.Src)
	req.u = u
	if err != nil {
		return nil, &getError{true, err}
	}

	// We have magic query parameters that we use to signal different features
	q := req.u.Query()

	// Determine if we have an archive type
	archiveV := q.Get("archive")
	if archiveV != "" {
		// Delete the paramter since it is a magic parameter we don't
		// want to pass on to the Getter
		q.Del("archive")
		req.u.RawQuery = q.Encode()

		// If we can parse the value as a bool and it is false, then
		// set the archive to "-" which should never map to a decompressor
		if b, err := strconv.ParseBool(archiveV); err == nil && !b {
			archiveV = "-"
		}
	} else {
		// We don't appear to... but is it part of the filename?
		matchingLen := 0
		for k := range c.Decompressors {
			if strings.HasSuffix(req.u.Path, "."+k) && len(k) > matchingLen {
				archiveV = k
				matchingLen = len(k)
			}
		}
	}

	// If we have a decompressor, then we need to change the destination
	// to download to a temporary path. We unarchive this into the final,
	// real path.
	var decompressDst string
	var decompressDir bool
	decompressor := c.Decompressors[archiveV]
	if decompressor != nil {
		// Create a temporary directory to store our archive. We delete
		// this at the end of everything.
		td, err := ioutil.TempDir("", "getter")
		if err != nil {
			return nil, &getError{true, fmt.Errorf(
				"Error creating temporary directory for archive: %s", err)}
		}
		defer os.RemoveAll(td)

		// Swap the download directory to be our temporary path and
		// store the old values.
		decompressDst = req.Dst
		decompressDir = req.Mode != ModeFile
		req.Dst = filepath.Join(td, "archive")
		req.Mode = ModeFile
	}

	// Determine checksum if we have one
	checksum, err := c.GetChecksum(ctx, req)
	if err != nil {
		return nil, &getError{true, fmt.Errorf("invalid checksum: %s", err)}
	}

	// Delete the query parameter if we have it.
	q.Del("checksum")
	req.u.RawQuery = q.Encode()

	if req.Mode == ModeAny {
		// Ask the getter which client mode to use
		req.Mode, err = g.Mode(ctx, req.u)
		if err != nil {
			return nil, &getError{false, err}
		}

		// Destination is the base name of the URL path in "any" mode when
		// a file source is detected.
		if req.Mode == ModeFile {
			filename := filepath.Base(req.u.Path)

			// Determine if we have a custom file name
			if v := q.Get("filename"); v != "" {
				// Delete the query parameter if we have it.
				q.Del("filename")
				req.u.RawQuery = q.Encode()

				filename = v
			}

			req.Dst = filepath.Join(req.Dst, filename)
		}
	}

	// If we're not downloading a directory, then just download the file
	// and return.
	if req.Mode == ModeFile {
		getFile := true
		if checksum != nil {
			if err := checksum.Checksum(req.Dst); err == nil {
				// don't get the file if the checksum of dst is correct
				getFile = false
			}
		}
		if getFile {
			if err := g.GetFile(ctx, req); err != nil {
				return nil, &getError{false, err}
			}

			if checksum != nil {
				if err := checksum.Checksum(req.Dst); err != nil {
					return nil, &getError{true, err}
				}
			}
		}

		if decompressor != nil {
			// We have a decompressor, so decompress the current destination
			// into the final destination with the proper mode.
			err := decompressor.Decompress(decompressDst, req.Dst, decompressDir)
			if err != nil {
				return nil, &getError{true, err}
			}

			// Swap the information back
			req.Dst = decompressDst
			if decompressDir {
				req.Mode = ModeAny
			} else {
				req.Mode = ModeFile
			}
		}

		// We check the dir value again because it can be switched back
		// if we were unarchiving. If we're still only Get-ing a file, then
		// we're done.
		if req.Mode == ModeFile {
			return &GetResult{req.Dst}, nil
		}
	}

	// If we're at this point we're either downloading a directory or we've
	// downloaded and unarchived a directory and we're just checking subdir.
	// In the case we have a decompressor we don't Get because it was Get
	// above.
	if decompressor == nil {
		// If we're getting a directory, then this is an error. You cannot
		// checksum a directory. TODO: test
		if checksum != nil {
			return nil, &getError{true, fmt.Errorf(
				"checksum cannot be specified for directory download")}
		}

		// We're downloading a directory, which might require a bit more work
		// if we're specifying a subdir.
		if err := g.Get(ctx, req); err != nil {
			return nil, &getError{false, err}
		}
	}

	// If we have a subdir, copy that over
	if req.subDir != "" {
		if err := os.RemoveAll(req.realDst); err != nil {
			return nil, &getError{true, err}
		}
		if err := os.MkdirAll(req.realDst, 0755); err != nil {
			return nil, &getError{true, err}
		}

		// Process any globs
		subDir, err := SubdirGlob(req.Dst, req.subDir)
		if err != nil {
			return nil, &getError{true, err}
		}

		err = copyDir(ctx, req.realDst, subDir, false)
		if err != nil {
			return nil, &getError{false, err}
		}
		return &GetResult{req.realDst}, nil
	}

	return &GetResult{req.Dst}, nil

}

func (c *Client) checkArchive(req *Request) string {
	q := req.u.Query()
	archiveV := q.Get("archive")
	if archiveV != "" {
		// Delete the paramter since it is a magic parameter we don't
		// want to pass on to the Getter
		q.Del("archive")
		req.u.RawQuery = q.Encode()

		// If we can parse the value as a bool and it is false, then
		// set the archive to "-" which should never map to a decompressor
		if b, err := strconv.ParseBool(archiveV); err == nil && !b {
			archiveV = "-"
		}
	}
	if archiveV == "" {
		// We don't appear to... but is it part of the filename?
		matchingLen := 0
		for k := range c.Decompressors {
			if strings.HasSuffix(req.u.Path, "."+k) && len(k) > matchingLen {
				archiveV = k
				matchingLen = len(k)
			}
		}
	}
	return archiveV
}
