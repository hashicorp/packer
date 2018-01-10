package common

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// PackerKeyEnv is used to specify the key interval (delay) between keystrokes
// sent to the VM, typically in boot commands. This is to prevent host CPU
// utilization from causing key presses to be skipped or repeated incorrectly.
const PackerKeyEnv = "PACKER_KEY_INTERVAL"

// PackerKeyDefault 100ms is appropriate for shared build infrastructure while a
// shorter delay (e.g. 10ms) can be used on a workstation. See PackerKeyEnv.
const PackerKeyDefault = 100 * time.Millisecond

// ScrubConfig is a helper that returns a string representation of
// any struct with the given values stripped out.
func ScrubConfig(target interface{}, values ...string) string {
	conf := fmt.Sprintf("Config: %+v", target)
	for _, value := range values {
		if value == "" {
			continue
		}
		conf = strings.Replace(conf, value, "<Filtered>", -1)
	}
	return conf
}

// ChooseString returns the first non-empty value.
func ChooseString(vals ...string) string {
	for _, el := range vals {
		if el != "" {
			return el
		}
	}

	return ""
}

// SupportedURL verifies that the url passed is actually supported or not
// This will also validate that the protocol is one that's actually implemented.
func SupportedURL(u *url.URL) bool {
	// url.Parse shouldn't return nil except on error....but it can.
	if u == nil {
		return false
	}

	// build a dummy NewDownloadClient since this is the only place that valid
	// protocols are actually exposed.
	cli := NewDownloadClient(&DownloadConfig{})

	// Iterate through each downloader to see if a protocol was found.
	ok := false
	for scheme, _ := range cli.config.DownloaderMap {
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
	var result string

	// Fix the url if it's using bad characters commonly mistaken with a path.
	original = filepath.ToSlash(original)

	// Check to see that this is a parseable URL with a scheme. If so, then just pass it through.
	if u, err := url.Parse(original); err == nil && u.Scheme != "" && u.Host != "" {
		return filepath.ToSlash(original), nil
	}

	// Since it's not a url, this might be a path. So, check that the file exists,
	// then make it an absolute path so we can make a proper uri.
	if _, err := os.Stat(original); err == nil {
		result, err = filepath.Abs(filepath.FromSlash(original))
		if err != nil {
			return "", err
		}

		result, err = filepath.EvalSymlinks(result)
		if err != nil {
			return "", err
		}

		result = filepath.Clean(result)
		result = filepath.ToSlash(result)

		// We have no idea what this might be, so we'll leave it as is.
	} else {
		result = filepath.ToSlash(original)
	}

	// We should have a path that can just turn into a file:// scheme'd url.
	return fmt.Sprintf("file://%s", result), nil
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
	if !SupportedURL(u) {
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

	fileURL, _ := url.Parse(original)
	fileExists := false

	if fileURL.Scheme == "file" {
		// on windows, correct URI is file:///c:/blah/blah.iso.
		// url.Parse will pull out the scheme "file://" and leave the path as
		// "/c:/blah/blah/iso".  Here we remove this forward slash on absolute
		// Windows file URLs before processing
		// see https://blogs.msdn.microsoft.com/ie/2006/12/06/file-uris-in-windows/
		// for more info about valid windows URIs
		filePath := fileURL.Path
		if runtime.GOOS == "windows" && len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}
		_, err := os.Stat(filePath)
		if err != nil {
			return fileExists
		} else {
			fileExists = true
		}
	}
	return fileExists
}
