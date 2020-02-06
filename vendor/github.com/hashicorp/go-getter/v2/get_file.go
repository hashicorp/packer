package getter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

// FileGetter is a Getter implementation that will download a module from
// a file scheme.
type FileGetter struct {
	getter
}

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
