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
			original = "file://" + filepath.ToSlash(original)
		}
	}

	// XXX: The validation here is later re-parsed in common/download.go and
	//      thus any modifications here must remain consistent over there too.
	uri, err := url.Parse(original)
	if err != nil {
		return "", err
	}

	if uri.Scheme == "" {
		uri.Scheme = "file"
	}

	const UNCPrefix = string(os.PathSeparator)+string(os.PathSeparator)
	if uri.Scheme == "file" {
		var ospath string	// os-formatted pathname
		if runtime.GOOS == "windows" {
			// Move any extra path components that were mis-parsed into the Host
			// field back into the uri.Path field
			if len(uri.Host) >= len(UNCPrefix) && uri.Host[:len(UNCPrefix)] == UNCPrefix {
				idx := strings.Index(uri.Host[len(UNCPrefix):], string(os.PathSeparator))
				if idx > -1 {
					uri.Path = filepath.ToSlash(uri.Host[idx+len(UNCPrefix):]) + uri.Path
					uri.Host = uri.Host[:idx+len(UNCPrefix)]
				}
			}
			// Now all we need to do to convert the uri to a platform-specific path
			// is to trade it's slashes for some os.PathSeparator ones.
			ospath = uri.Host + filepath.FromSlash(uri.Path)

		} else {
			// Since we're already using sane paths on a sane platform, anything in
			// uri.Host can be assumed that the user is describing a relative uri.
			// This means that if we concatenate it with uri.Path, the filepath
			// transform will still open the file correctly.
			//     i.e. file://localdirectory/filename -> localdirectory/filename
			ospath = uri.Host + uri.Path
		}
		// Only do the filepath transformations if the file appears
		// to actually exist. We don't do it on windows, because EvalSymlinks
		// won't understand how to handle UNC paths and other Windows-specific minutae.
		if _, err := os.Stat(ospath); err == nil && runtime.GOOS != "windows" {
			ospath, err = filepath.Abs(ospath)
			if err != nil {
				return "", err
			}

			ospath, err = filepath.EvalSymlinks(ospath)
			if err != nil {
				return "", err
			}

			ospath = filepath.Clean(ospath)
		}

		// now that ospath was normalized and such..
		if runtime.GOOS == "windows" {
			uri.Host = ""
			// Check to see if our ospath is unc-prefixed, and if it is then split
			// the UNC host into uri.Host, leaving the rest in ospath.
			// This way, our UNC-uri is protected from injury in the call to uri.String()
			if len(ospath) >= len(UNCPrefix) && ospath[:len(UNCPrefix)] == UNCPrefix {
				idx := strings.Index(ospath[len(UNCPrefix):], string(os.PathSeparator))
				if idx > -1 {
					uri.Host = ospath[:len(UNCPrefix)+idx]
					ospath = ospath[len(UNCPrefix)+idx:]
				}
			}
			// Restore the uri by re-transforming our os-formatted path
			uri.Path = filepath.ToSlash(ospath)
		} else {
			uri.Host = ""
			uri.Path = filepath.ToSlash(ospath)
		}
	}

	// Make sure it is lowercased
	uri.Scheme = strings.ToLower(uri.Scheme)

	// Verify that the scheme is something we support in our common downloader.
	supported := []string{"file", "http", "https"}
	found := false
	for _, s := range supported {
		if uri.Scheme == s {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("Unsupported URL scheme: %s", uri.Scheme)
	}

	// explicit check to see if we need to manually replace the uri host with a UNC one
	if runtime.GOOS == "windows" && uri.Scheme == "file" {
		if len(uri.Host) >= len(UNCPrefix) && uri.Host[:len(UNCPrefix)] == UNCPrefix {
			escapedHost := url.QueryEscape(uri.Host)
			return strings.Replace(uri.String(), escapedHost, uri.Host, 1), nil
		}
	}
	return uri.String(), nil
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
