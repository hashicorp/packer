package common

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer"
)

// PackerKeyEnv is used to specify the key interval (delay) between keystrokes
// sent to the VM, typically in boot commands. This is to prevent host CPU
// utilization from causing key presses to be skipped or repeated incorrectly.
const PackerKeyEnv = "PACKER_KEY_INTERVAL"

// PackerKeyDefault 100ms is appropriate for shared build infrastructure while a
// shorter delay (e.g. 10ms) can be used on a workstation. See PackerKeyEnv.
const PackerKeyDefault = 100 * time.Millisecond

// ChooseString returns the first non-empty value.
func ChooseString(vals ...string) string {
	for _, el := range vals {
		if el != "" {
			return el
		}
	}

	return ""
}

// SupportedProtocol verifies that the url passed is actually supported or not
// This will also validate that the protocol is one that's actually implemented.
func SupportedProtocol(u *url.URL) bool {
	// url.Parse shouldn't return nil except on error....but it can.
	if u == nil {
		return false
	}

	// build a dummy NewDownloadClient since this is the only place that valid
	// protocols are actually exposed.
	cli := NewDownloadClient(&DownloadConfig{}, new(packer.NoopProgressBar))

	// Iterate through each downloader to see if a protocol was found.
	ok := false
	for scheme := range cli.config.DownloaderMap {
		if strings.ToLower(u.Scheme) == strings.ToLower(scheme) {
			ok = true
		}
	}
	return ok
}

// DownloadableURL processes a URL that may also be a file path and returns
// a completely valid URL representing the requested file. For example,
// the original URL might be "local/file.iso" which isn't a valid URL,
// and so DownloadableURL will return "file://local/file.iso"
// No other transformations are done to the path.
func DownloadableURL(original string) (string, error) {
	var absPrefix, result string

	absPrefix = ""
	if runtime.GOOS == "windows" {
		absPrefix = "/"
	}

	// Check that the user specified a UNC path, and promote it to an smb:// uri.
	if strings.HasPrefix(original, "\\\\") && len(original) > 2 && original[2] != '?' {
		result = filepath.ToSlash(original[2:])
		return fmt.Sprintf("smb://%s", result), nil
	}

	// Fix the url if it's using bad characters commonly mistaken with a path.
	original = filepath.ToSlash(original)

	// Check to see that this is a parseable URL with a scheme and a host.
	// If so, then just pass it through.
	if u, err := url.Parse(original); err == nil && u.Scheme != "" && u.Host != "" {
		return original, nil
	}

	// If it's a file scheme, then convert it back to a regular path so the next
	// case which forces it to an absolute path, will correct it.
	if u, err := url.Parse(original); err == nil && strings.ToLower(u.Scheme) == "file" {
		original = u.Path
	}

	// If we're on Windows and we start with a slash, then this absolute path
	// is wrong. Fix it up, so the next case can figure out the absolute path.
	if rpath := strings.SplitN(original, "/", 2); rpath[0] == "" && runtime.GOOS == "windows" {
		result = rpath[1]
	} else {
		result = original
	}

	// Since we should be some kind of path (relative or absolute), check
	// that the file exists, then make it an absolute path so we can return an
	// absolute uri.
	if _, err := os.Stat(result); err == nil {
		result, err = filepath.Abs(filepath.FromSlash(result))
		if err != nil {
			return "", err
		}

		result, err = filepath.EvalSymlinks(result)
		if err != nil {
			return "", err
		}

		result = filepath.Clean(result)
		return fmt.Sprintf("file://%s%s", absPrefix, filepath.ToSlash(result)), nil
	}

	// Otherwise, check if it was originally an absolute path, and fix it if so.
	if strings.HasPrefix(original, "/") {
		return fmt.Sprintf("file://%s%s", absPrefix, result), nil
	}

	// Anything left should be a non-existent relative path. So fix it up here.
	result = filepath.ToSlash(filepath.Clean(result))
	return fmt.Sprintf("file://./%s", result), nil
}

// Force the parameter into a url. This will transform the parameter into
// a proper url, removing slashes, adding the proper prefix, etc.
func ValidatedURL(original string) (string, error) {

	// See if the user failed to give a url
	if ok, _ := regexp.MatchString("(?m)^[^[:punct:]]+://", original); !ok {

		// So since no magic was found, this must be a path.
		result, err := DownloadableURL(original)
		if err == nil {
			return ValidatedURL(result)
		}

		return "", err
	}

	// Verify that the url is parseable...just in case.
	u, err := url.Parse(original)
	if err != nil {
		return "", err
	}

	// We should now have a url, so verify that it's a protocol we support.
	if !SupportedProtocol(u) {
		return "", fmt.Errorf("Unsupported protocol scheme! (%#v)", u)
	}

	// We should now have a properly formatted and supported url
	return u.String(), nil
}

// FileExistsLocally takes the URL output from DownloadableURL, and determines
// whether it is present on the file system.
// example usage:
//
// myFile, err = common.DownloadableURL(c.SourcePath)
// ...
// fileExists := common.StatURL(myFile)
// possible output:
// true -- should occur if the file is present, or if the file is not present,
// but is not supposed to be (e.g. the schema is http://, not file://)
// false -- should occur if there was an error stating the file, so the
// file is not present when it should be.

func FileExistsLocally(original string) bool {
	// original should be something like file://C:/my/path.iso
	u, _ := url.Parse(original)

	// First create a dummy downloader so we can figure out which
	// protocol to use.
	cli := NewDownloadClient(&DownloadConfig{}, new(packer.NoopProgressBar))
	d, ok := cli.config.DownloaderMap[u.Scheme]
	if !ok {
		return false
	}

	// Check to see that it's got a Local way of doing things.
	local, ok := d.(LocalDownloader)
	if !ok {
		return true // XXX: Remote URLs short-circuit this logic.
	}

	// Figure out where we're at.
	wd, err := os.Getwd()
	if err != nil {
		return false
	}

	// Now figure out the real path to the file.
	realpath, err := local.toPath(wd, *u)
	if err != nil {
		return false
	}

	// Finally we can seek the truth via os.Stat.
	_, err = os.Stat(realpath)
	if err != nil {
		return false
	}
	return true
}
