package getter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
)

// FileGetter is a Getter implementation that will download a module from
// a file scheme.
type FileGetter struct{}

func (g *FileGetter) Mode(ctx context.Context, u *url.URL) (Mode, error) {
	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	// Check if the source is a directory.
	if fi.IsDir() {
		return ModeDir, nil
	}

	return ModeFile, nil
}

func (g *FileGetter) Get(ctx context.Context, req *Request) error {
	path := req.u.Path
	if req.u.RawPath != "" {
		path = req.u.RawPath
	}

	// The source path must exist and be a directory to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if !fi.IsDir() {
		return fmt.Errorf("source path must be a directory")
	}

	fi, err := os.Lstat(req.Dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if req.Inplace {
		req.Dst = path
		return nil
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		mode := fi.Mode()
		if mode&os.ModeSymlink == 0 {
			return fmt.Errorf("destination exists and is not a symlink")
		}

		// Remove the destination
		if err := os.Remove(req.Dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(req.Dst), 0755); err != nil {
		return err
	}

	return SymlinkAny(path, req.Dst)
}

func (g *FileGetter) GetFile(ctx context.Context, req *Request) error {
	path := req.u.Path
	if req.u.RawPath != "" {
		path = req.u.RawPath
	}

	// The source path must exist and be a file to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if fi.IsDir() {
		return fmt.Errorf("source path must be a file")
	}

	if req.Inplace {
		req.Dst = path
		return nil
	}

	_, err := os.Lstat(req.Dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		// Remove the destination
		if err := os.Remove(req.Dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(req.Dst), 0755); err != nil {
		return err
	}

	// If we're not copying, just symlink and we're done
	if !req.Copy {
		if err = os.Symlink(path, req.Dst); err == nil {
			return err
		}
		lerr, ok := err.(*os.LinkError)
		if !ok {
			return err
		}
		switch lerr.Err {
		case ErrUnauthorized:
			// On windows this  means we don't have
			// symlink privilege, let's
			// fallback to a copy to avoid an error.
			break
		default:
			return err
		}
	}

	// Copy
	srcF, err := os.Open(path)
	if err != nil {
		return err
	}
	defer srcF.Close()

	dstF, err := os.Create(req.Dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	_, err = Copy(ctx, dstF, srcF)
	return err
}

func (g *FileGetter) Detect(req *Request) (bool, error) {
	var src, pwd string
	src = req.Src
	pwd = req.Pwd
	if len(src) == 0 {
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

	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" {
		if isForcedGetter {
			// Is the Forced getter and source is a valid url
			return true, nil
		}
		if g.validScheme(u.Scheme) {
			return true, nil
		}
		if !(runtime.GOOS == "windows" && len(u.Scheme) == 1) {
			return false, nil
		}
		// For windows, we try to get the artifact
		// if it has a non valid scheme with 1 char
		// e.g. C:/foo/bar for other cases a prefix file:// is necessary
	}

	src, ok, err := new(FileDetector).Detect(src, pwd)
	if err != nil {
		return ok, err
	}
	if ok {
		req.Src = src
		return ok, nil
	}

	return true, nil
}

func (g *FileGetter) validScheme(scheme string) bool {
	return scheme == "file"
}
