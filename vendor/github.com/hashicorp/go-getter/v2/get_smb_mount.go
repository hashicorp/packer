package getter

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
)

// SmbMountGetter is a Getter implementation that will download an artifact from
// a shared folder using the file system using FileGetter implementation.
// For Unix and MacOS users, the Getter will look for usual system specific mount paths such as:
// /Volumes/ for MacOS
// /run/user/1000/gvfs/smb-share:server=<hostIP>,share=<path> for Unix
type SmbMountGetter struct{}

func (g *SmbMountGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	if u.Host == "" || u.Path == "" {
		return 0, new(smbPathError)
	}

	prefix, path := g.findPrefixAndPath(u)
	u.Path = prefix + path

	return new(FileGetter).Mode(ctx, u)
}

func (g *SmbMountGetter) Get(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	prefix, path := g.findPrefixAndPath(req.u)
	req.u.Path = prefix + path

	return new(FileGetter).Get(ctx, req)
}

func (g *SmbMountGetter) GetFile(ctx context.Context, req *Request) error {
	if req.u.Host == "" || req.u.Path == "" {
		return new(smbPathError)
	}

	prefix, path := g.findPrefixAndPath(req.u)
	req.u.Path = prefix + path

	return new(FileGetter).GetFile(ctx, req)
}

func (g *SmbMountGetter) findPrefixAndPath(u *url.URL) (string, string) {
	var prefix, path string
	switch runtime.GOOS {
	case "windows":
		prefix = string(os.PathSeparator) + string(os.PathSeparator)
		path = filepath.Join(u.Host, u.Path)
	case "darwin":
		prefix = string(os.PathSeparator)
		path = filepath.Join("Volumes", u.Path)
	}
	return prefix, path
}

func (g *SmbMountGetter) Detect(req *Request) (bool, error) {
	if runtime.GOOS == "linux" {
		// Linux has the smbclient command which is a safer approach to retrieve an artifact from a samba shared folder.
		// Therefore, this should be used instead of looking in the file system.
		return false, nil
	}
	if len(req.Src) == 0 {
		return false, nil
	}

	if req.Forced != "" {
		// There's a getter being Forced
		if !g.validScheme(req.Forced) {
			// Current getter is not the Forced one
			// Don't use it to try to download the artifact
			return false, nil
		}
	}
	isForcedGetter := req.Forced != "" && g.validScheme(req.Forced)

	u, err := url.Parse(req.Src)
	if err == nil && u.Scheme != "" {
		if isForcedGetter {
			// Is the Forced getter and source is a valid url
			return true, nil
		}
		if g.validScheme(u.Scheme) {
			return true, nil
		}
		// Valid url with a scheme that is not valid for current getter
		return false, nil
	}

	return false, nil
}

func (g *SmbMountGetter) validScheme(scheme string) bool {
	return scheme == "smb"
}
