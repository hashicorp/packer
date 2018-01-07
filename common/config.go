package common

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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

	// Verify that the scheme is something we support in our common downloader.
	supported := []string{"file", "http", "https", "smb"}
	found := false
	for _, s := range supported {
		if strings.HasPrefix(strings.ToLower(original), s+"://") {
			found = true
			break
		}
	}

	// If it's properly prefixed with something we support, then we don't need
	//	to make it a uri.
	if found {
		original = filepath.ToSlash(original)

		// make sure that it can be parsed though..
		uri, err := url.Parse(original)
		if err != nil {
			return "", err
		}

		uri.Scheme = strings.ToLower(uri.Scheme)

		return uri.String(), nil
	}

	// If the file exists, then make it an absolute path
	_, err := os.Stat(original)
	if err == nil {
		original, err = filepath.Abs(filepath.FromSlash(original))
		if err != nil {
			return "", err
		}

		original, err = filepath.EvalSymlinks(original)
		if err != nil {
			return "", err
		}

		original = filepath.Clean(original)
		original = filepath.ToSlash(original)
	}

	// Since it wasn't properly prefixed, let's make it into a well-formed
	//	file:// uri.

	return "file://" + original, nil
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
