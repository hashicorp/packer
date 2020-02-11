package getter

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
	safetemp "github.com/hashicorp/go-safetemp"
)

// Client is a client for downloading things.
//
// Top-level functions such as Get are shortcuts for interacting with a client.
// Using a client directly allows more fine-grained control over how downloading
// is done, as well as customizing the protocols supported.
type Client struct {

	// Detectors is the list of detectors that are tried on the source.
	// If this is nil, then the default Detectors will be used.
	Detectors []Detector

	// Decompressors is the map of decompressors supported by this client.
	// If this is nil, then the default value is the Decompressors global.
	Decompressors map[string]Decompressor

	// Getters is the map of protocols supported by this client. If this
	// is nil, then the default Getters variable will be used.
	Getters map[string]Getter
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

	var err error
	req.Src, err = Detect(req.Src, req.Pwd, c.Detectors)
	if err != nil {
		return nil, err
	}

	var force string
	// Determine if we have a forced protocol, i.e. "git::http://..."
	force, req.Src = getForcedGetter(req.Src)

	// If there is a subdir component, then we download the root separately
	// and then copy over the proper subdir.
	var realDst, subDir string
	req.Src, subDir = SourceDirSubdir(req.Src)
	if subDir != "" {
		td, tdcloser, err := safetemp.Dir("", "getter")
		if err != nil {
			return nil, err
		}
		defer tdcloser.Close()

		realDst = req.Dst
		req.Dst = td
	}

	req.u, err = urlhelper.Parse(req.Src)
	if err != nil {
		return nil, err
	}
	if force == "" {
		force = req.u.Scheme
	}

	g, ok := c.Getters[force]
	if !ok {
		return nil, fmt.Errorf(
			"download not supported for scheme '%s'", force)
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
			return nil, fmt.Errorf(
				"Error creating temporary directory for archive: %s", err)
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
	checksum, err := c.extractChecksum(ctx, req.u)
	if err != nil {
		return nil, fmt.Errorf("invalid checksum: %s", err)
	}

	// Delete the query parameter if we have it.
	q.Del("checksum")
	req.u.RawQuery = q.Encode()

	if req.Mode == ModeAny {
		// Ask the getter which client mode to use
		req.Mode, err = g.Mode(ctx, req.u)
		if err != nil {
			return nil, err
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
			if err := checksum.checksum(req.Dst); err == nil {
				// don't get the file if the checksum of dst is correct
				getFile = false
			}
		}
		if getFile {
			err := g.GetFile(ctx, req)
			if err != nil {
				return nil, err
			}

			if checksum != nil {
				if err := checksum.checksum(req.Dst); err != nil {
					return nil, err
				}
			}
		}

		if decompressor != nil {
			// We have a decompressor, so decompress the current destination
			// into the final destination with the proper mode.
			err := decompressor.Decompress(decompressDst, req.Dst, decompressDir)
			if err != nil {
				return nil, err
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
			return nil, fmt.Errorf(
				"checksum cannot be specified for directory download")
		}

		// We're downloading a directory, which might require a bit more work
		// if we're specifying a subdir.
		err := g.Get(ctx, req)
		if err != nil {
			err = fmt.Errorf("error downloading '%s': %s", req.Src, err)
			return nil, err
		}
	}

	// If we have a subdir, copy that over
	if subDir != "" {
		if err := os.RemoveAll(realDst); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(realDst, 0755); err != nil {
			return nil, err
		}

		// Process any globs
		subDir, err := SubdirGlob(req.Dst, subDir)
		if err != nil {
			return nil, err
		}

		return &GetResult{realDst}, copyDir(ctx, realDst, subDir, false)
	}

	return &GetResult{req.Dst}, nil
}
