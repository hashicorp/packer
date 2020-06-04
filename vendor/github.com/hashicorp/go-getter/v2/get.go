// getter is a package for downloading files or directories from a variety of
// protocols.
//
// getter is unique in its ability to download both directories and files.
// It also detects certain source strings to be protocol-specific URLs. For
// example, "github.com/hashicorp/go-getter/v2" would turn into a Git URL and
// use the Git protocol.
//
// Protocols and detectors are extensible.
//
// To get started, see Client.
package getter

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"syscall"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
)

// Getter defines the interface that schemes must implement to download
// things.
type Getter interface {
	// Get downloads the given URL into the given directory. This always
	// assumes that we're updating and gets the latest version that it can.
	//
	// The directory may already exist (if we're updating). If it is in a
	// format that isn't understood, an error should be returned. Get shouldn't
	// simply nuke the directory.
	Get(context.Context, *Request) error

	// GetFile downloads the give URL into the given path. The URL must
	// reference a single file. If possible, the Getter should check if
	// the remote end contains the same file and no-op this operation.
	GetFile(context.Context, *Request) error

	// Mode returns the mode based on the given URL. This is used to
	// allow clients to let the getters decide which mode to use.
	Mode(context.Context, *url.URL) (Mode, error)

	// Detect detects whether the Request.Src matches a known pattern to
	// turn it into a proper URL, and also transforms and update Request.Src
	// when necessary.
	// The Getter must validate if the Request.Src is a valid URL
	// with a valid scheme for the Getter, and also check if the
	// current Getter is the forced one and return true if that's the case.
	Detect(*Request) (bool, error)
}

// Getters is the mapping of scheme to the Getter implementation that will
// be used to get a dependency.
var Getters []Getter

// forcedRegexp is the regular expression that finds Forced getters. This
// syntax is schema::url, example: git::https://foo.com
var forcedRegexp = regexp.MustCompile(`^([A-Za-z0-9]+)::(.+)$`)

// httpClient is the default client to be used by HttpGetters.
var httpClient = cleanhttp.DefaultClient()

var DefaultClient = &Client{
	Getters:       Getters,
	Decompressors: Decompressors,
}

func init() {
	httpGetter := &HttpGetter{
		Netrc: true,
	}

	// The order of the Getters in the list may affect the result
	// depending if the Request.Src is detected as valid by multiple getters
	Getters = []Getter{
		&GitGetter{[]Detector{
			new(GitHubDetector),
			new(GitDetector),
			new(BitBucketDetector),
		},
		},
		new(HgGetter),
		new(SmbClientGetter),
		new(SmbMountGetter),
		httpGetter,
		new(FileGetter),
	}
}

// Get downloads the directory specified by src into the folder specified by
// dst. If dst already exists, Get will attempt to update it.
//
// src is a URL, whereas dst is always just a file path to a folder. This
// folder doesn't need to exist. It will be created if it doesn't exist.
func Get(ctx context.Context, dst, src string) (*GetResult, error) {
	req := &Request{
		Src:  src,
		Dst:  dst,
		Mode: ModeDir,
	}
	return DefaultClient.Get(ctx, req)
}

// GetAny downloads a URL into the given destination. Unlike Get or
// GetFile, both directories and files are supported.
//
// dst must be a directory. If src is a file, it will be downloaded
// into dst with the basename of the URL. If src is a directory or
// archive, it will be unpacked directly into dst.
func GetAny(ctx context.Context, dst, src string) (*GetResult, error) {
	req := &Request{
		Src:  src,
		Dst:  dst,
		Mode: ModeAny,
	}
	return DefaultClient.Get(ctx, req)
}

// GetFile downloads the file specified by src into the path specified by
// dst.
func GetFile(ctx context.Context, dst, src string) (*GetResult, error) {
	req := &Request{
		Src:  src,
		Dst:  dst,
		Mode: ModeFile,
	}
	return DefaultClient.Get(ctx, req)
}

// getRunCommand is a helper that will run a command and capture the output
// in the case an error happens.
func getRunCommand(cmd *exec.Cmd) error {
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if err == nil {
		return nil
	}
	if exiterr, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return fmt.Errorf(
				"%s exited with %d: %s",
				cmd.Path,
				status.ExitStatus(),
				buf.String())
		}
	}

	return fmt.Errorf("error running %s: %s", cmd.Path, buf.String())
}

// getForcedGetter takes a source and returns the tuple of the forced
// getter and the raw URL (without the force syntax).
// For example "git::https://...". returns "git" "https://".
func getForcedGetter(src string) (string, string) {
	var forced string
	if ms := forcedRegexp.FindStringSubmatch(src); ms != nil {
		forced = ms[1]
		src = ms[2]
	}

	return forced, src
}
