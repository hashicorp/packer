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
	fmt.Printf("Swampy: user input was %s\n", original)
	if runtime.GOOS == "windows" {
		// If the distance to the first ":" is just one character, assume
		// we're dealing with a drive letter and thus a file path.
		idx := strings.Index(original, ":")
		if idx == 1 {
			original = "file:///" + original
		}
	}

	url, err := url.Parse(original)
	if err != nil {
		return "", err
	}

	if url.Scheme == "" {
		url.Scheme = "file"
	}

	if url.Scheme == "file" {
		// Windows file handling is all sorts of tricky...
		if runtime.GOOS == "windows" {
			// If the path is using Windows-style slashes, URL parses
			// it into the host field.
			if url.Path == "" && strings.Contains(url.Host, `\`) {
				url.Path = url.Host
				url.Host = ""
			}
		}

		// Only do the filepath transformations if the file appears
		// to actually exist.
		if _, err := os.Stat(url.Path); err == nil {
			url.Path, err = filepath.Abs(url.Path)
			if err != nil {
				return "", err
			}

			url.Path, err = filepath.EvalSymlinks(url.Path)
			if err != nil {
				return "", err
			}

			// url.Path = filepath.Clean(url.Path)
		}

		if runtime.GOOS == "windows" {
			// Also replace all backslashes with forwardslashes since Windows
			// users are likely to do this but the URL should actually only
			// contain forward slashes.
			url.Path = strings.Replace(url.Path, `\`, `/`, -1)
		}
	}

	// Make sure it is lowercased
	url.Scheme = strings.ToLower(url.Scheme)

	// Verify that the scheme is something we support in our common downloader.
	supported := []string{"file", "http", "https"}
	found := false
	for _, s := range supported {
		if url.Scheme == s {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("Unsupported URL scheme: %s", url.Scheme)
	}
	fmt.Printf("Swampy: parsed string after DownloadableURL is %s\n", url.String())
	return url.String(), nil
}

// FileExistsLocally takes the URL output from DownloadableURL, and determines
// whether it is present on the file system.
// example usage:
//
// myFile, err = common.DownloadableURL(c.SourcePath)
// ...
// fileExists, err := common.StatURL(myFile)
// possible output:
// true, nil -- should occur if the file is present
// false, nil -- should occur if the file is not present, but is not supposed to
// be (e.g. the schema is http://, not file://)
// true, error -- shouldn't occur ever
// false, error -- should occur if there was an error stating the file, so the
// file is not present when it should be.

func FileExistsLocally(original string) (bool, error) {
	// original should be something like file://C:/my/path.iso
	// on windows, c drive will be parsed as host if it's file://c instead of file:///c
	prefix = "file://"
	filePath = strings.Replace(original, prefix, "", 1)
	fmt.Printf("Swampy: original is %s\n", original)
	fmt.Printf("Swampy: filePath is %#v\n", filePath)

	if fileURL.Scheme == "file" {
		_, err := os.Stat(filePath)
		if err != nil {
			err = fmt.Errorf("could not stat file: %s\n", err)
			return fileExists, err
		} else {
			fileExists = true
		}
	}
	return fileExists, nil
}
