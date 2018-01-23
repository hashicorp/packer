package common

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
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

// DownloadableURL processes a URL that may also be a file path and returns
// a completely valid URL. For example, the original URL might be "local/file.iso"
// which isn't a valid URL. DownloadableURL will return "file:///local/file.iso"
func DownloadableURL(original string) (string, error) {
	if runtime.GOOS == "windows" {
		// If the distance to the first ":" is just one character, assume
		// we're dealing with a drive letter and thus a file path.
		// prepend with "file:///"" now so that url.Parse won't accidentally
		// parse the drive letter into the url scheme.
		// See https://blogs.msdn.microsoft.com/ie/2006/12/06/file-uris-in-windows/
		// for more info about valid windows URIs
		idx := strings.Index(original, ":")
		if idx == 1 {
			original = "file:///" + original
		}
	}
	u, err := url.Parse(original)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" {
		u.Scheme = "file"
	}

	if u.Scheme == "file" {
		// Windows file handling is all sorts of tricky...
		if runtime.GOOS == "windows" {
			// If the path is using Windows-style slashes, URL parses
			// it into the host field.
			if u.Path == "" && strings.Contains(u.Host, `\`) {
				u.Path = u.Host
				u.Host = ""
			}
		}
		// Only do the filepath transformations if the file appears
		// to actually exist.
		if _, err := os.Stat(u.Path); err == nil {
			u.Path, err = filepath.Abs(u.Path)
			if err != nil {
				return "", err
			}

			u.Path, err = filepath.EvalSymlinks(u.Path)
			if err != nil {
				return "", err
			}

			u.Path = filepath.Clean(u.Path)
		}

		if runtime.GOOS == "windows" {
			// Also replace all backslashes with forwardslashes since Windows
			// users are likely to do this but the URL should actually only
			// contain forward slashes.
			u.Path = strings.Replace(u.Path, `\`, `/`, -1)
			// prepend absolute windows paths with "/" so that when we
			// compose u.String() below the outcome will be correct
			// file:///c/blah syntax; otherwise u.String() will only add
			// file:// which is not technically a correct windows URI
			if filepath.IsAbs(u.Path) && !strings.HasPrefix(u.Path, "/") {
				u.Path = "/" + u.Path
			}

		}
	}

	// Make sure it is lowercased
	u.Scheme = strings.ToLower(u.Scheme)

	// Verify that the scheme is something we support in our common downloader.
	supported := []string{"file", "http", "https"}
	found := false
	for _, s := range supported {
		if u.Scheme == s {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("Unsupported URL scheme: %s", u.Scheme)
	}
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
