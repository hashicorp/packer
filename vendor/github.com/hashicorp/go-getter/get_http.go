package getter

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HttpGetter is a Getter implementation that will download from an HTTP
// endpoint.
//
// For file downloads, HTTP is used directly.
//
// The protocol for downloading a directory from an HTTP endpoing is as follows:
//
// An HTTP GET request is made to the URL with the additional GET parameter
// "terraform-get=1". This lets you handle that scenario specially if you
// wish. The response must be a 2xx.
//
// First, a header is looked for "X-Terraform-Get" which should contain
// a source URL to download.
//
// If the header is not present, then a meta tag is searched for named
// "terraform-get" and the content should be a source URL.
//
// The source URL, whether from the header or meta tag, must be a fully
// formed URL. The shorthand syntax of "github.com/foo/bar" or relative
// paths are not allowed.
type HttpGetter struct {
	// Netrc, if true, will lookup and use auth information found
	// in the user's netrc file if available.
	Netrc bool

	// Client is the http.Client to use for Get requests.
	// This defaults to a cleanhttp.DefaultClient if left unset.
	Client *http.Client

	// Used for calculating percent progress
	totalSize       int64
	PercentComplete int
	dstFile         string
	Done            chan int64
}

func (g *HttpGetter) ClientMode(u *url.URL) (ClientMode, error) {
	if strings.HasSuffix(u.Path, "/") {
		return ClientModeDir, nil
	}
	return ClientModeFile, nil
}

func (g *HttpGetter) Get(dst string, u *url.URL) error {
	// Copy the URL so we can modify it
	var newU url.URL = *u
	u = &newU

	if g.Netrc {
		// Add auth from netrc if we can
		if err := addAuthFromNetrc(u); err != nil {
			return err
		}
	}

	if g.Client == nil {
		g.Client = httpClient
	}

	g.dstFile = dst

	// Add terraform-get to the parameter.
	q := u.Query()
	q.Add("terraform-get", "1")
	u.RawQuery = q.Encode()
	// Get the URL
	resp, err := g.Client.Get(u.String())
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	// Extract the source URL
	var source string
	if v := resp.Header.Get("X-Terraform-Get"); v != "" {
		source = v
	} else {
		source, err = g.parseMeta(resp.Body)
		if err != nil {
			return err
		}
	}
	if source == "" {
		return fmt.Errorf("no source URL was returned")
	}

	// extract the source file size
	g.totalSize = resp.ContentLength

	// If there is a subdir component, then we download the root separately
	// into a temporary directory, then copy over the proper subdir.
	source, subDir := SourceDirSubdir(source)
	if subDir == "" {
		return Get(dst, source)
	}
	// We have a subdir, time to jump some hoops
	return g.getSubdir(dst, source, subDir)
}

func (g *HttpGetter) GetFile(dst string, u *url.URL) error {
	g.dstFile = dst

	if g.Netrc {
		// Add auth from netrc if we can
		if err := addAuthFromNetrc(u); err != nil {
			return err
		}
	}

	if g.Client == nil {
		g.Client = httpClient
	}
	resp, err := g.Client.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	// extract the source file size
	g.totalSize = resp.ContentLength

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	go g.CalcDownloadPercent()

	g.Done = make(chan int64, 1)
	nwritten, err := io.Copy(f, resp.Body)

	// tell progress tracker we're done
	g.Done <- nwritten

	return err
}

// getSubdir downloads the source into the destination, but with
// the proper subdir.
func (g *HttpGetter) getSubdir(dst, source, subDir string) error {
	// Create a temporary directory to store the full source
	td, err := ioutil.TempDir("", "tf")
	if err != nil {
		return err
	}
	defer os.RemoveAll(td)

	// We have to create a subdirectory that doesn't exist for the file
	// getter to work.
	td = filepath.Join(td, "data")

	// Download that into the given directory
	if err := Get(td, source); err != nil {
		return err
	}

	// Process any globbing
	sourcePath, err := SubdirGlob(td, subDir)
	if err != nil {
		return err
	}

	// Make sure the subdir path actually exists
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf(
			"Error downloading %s: %s", source, err)
	}

	// Copy the subdirectory into our actual destination.
	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	// Make the final destination
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return copyDir(dst, sourcePath, false)
}

// parseMeta looks for the first meta tag in the given reader that
// will give us the source URL.
func (g *HttpGetter) parseMeta(r io.Reader) (string, error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charsetReader
	d.Strict = false
	var err error
	var t xml.Token
	for {
		t, err = d.Token()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return "", err
		}
		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, "body") {
			return "", nil
		}
		if e, ok := t.(xml.EndElement); ok && strings.EqualFold(e.Name.Local, "head") {
			return "", nil
		}
		e, ok := t.(xml.StartElement)
		if !ok || !strings.EqualFold(e.Name.Local, "meta") {
			continue
		}
		if attrValue(e.Attr, "name") != "terraform-get" {
			continue
		}
		if f := attrValue(e.Attr, "content"); f != "" {
			return f, nil
		}
	}
}

func (g *HttpGetter) GetProgress() int {
	return g.PercentComplete
}

func (g *HttpGetter) CalcDownloadPercent() {
	// stat file every n seconds to figure out the download progress
	var stop bool = false
	dstfile, err := os.Open(g.dstFile)
	defer dstfile.Close()

	if err != nil {
		log.Printf("couldn't open file for reading: %s", err)
		return
	}
	for {
		select {
		case <-g.Done:
			stop = true
		default:
			g.PercentComplete = CalcPercent(dstfile, g.totalSize)
		}

		if stop {
			break
		}
		// repeat check once per second
		time.Sleep(time.Second)
	}
}

// attrValue returns the attribute value for the case-insensitive key
// `name', or the empty string if nothing is found.
func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}

// charsetReader returns a reader for the given charset. Currently
// it only supports UTF-8 and ASCII. Otherwise, it returns a meaningful
// error which is printed by go get, so the user can find why the package
// wasn't downloaded if the encoding is not supported. Note that, in
// order to reduce potential errors, ASCII is treated as UTF-8 (i.e. characters
// greater than 0x7f are not rejected).
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "ascii":
		return input, nil
	default:
		return nil, fmt.Errorf("can't decode XML document using charset %q", charset)
	}
}
