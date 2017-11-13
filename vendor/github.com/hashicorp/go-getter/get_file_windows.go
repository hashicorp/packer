// +build windows

package getter

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (g *FileGetter) Get(dst string, u *url.URL) error {
	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	// The source path must exist and be a directory to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if !fi.IsDir() {
		return fmt.Errorf("source path must be a directory")
	}

	fi, err := os.Lstat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		mode := fi.Mode()
		if mode&os.ModeSymlink == 0 {
			return fmt.Errorf("destination exists and is not a symlink")
		}

		// Remove the destination
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	sourcePath := toBackslash(path)

	// Use mklink to create a junction point
	output, err := exec.Command("cmd", "/c", "mklink", "/J", dst, sourcePath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run mklink %v %v: %v %q", dst, sourcePath, err, output)
	}
	g.PercentComplete = 100
	return nil
}

func (g *FileGetter) GetFile(dst string, u *url.URL) error {
	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	// The source path must exist and be a directory to be usable.
	if fi, err := os.Stat(path); err != nil {
		return fmt.Errorf("source path error: %s", err)
	} else if fi.IsDir() {
		return fmt.Errorf("source path must be a file")
	}
	g.totalSize = fi.Size()

	_, err := os.Lstat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the destination already exists, it must be a symlink
	if err == nil {
		// Remove the destination
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// If we're not copying, just symlink and we're done
	if !g.Copy {
		g.PercentComplete = 100
		return os.Symlink(path, dst)
	}

	// Copy
	srcF, err := os.Open(path)
	if err != nil {
		return err
	}
	defer srcF.Close()

	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	// track copy progress
	g.Done = make(chan int64, 1)
	go g.CalcPercentComplete(dst)

	nwritten, err = io.Copy(dstF, srcF)
	g.Done <- nwritten
	return err
}

// toBackslash returns the result of replacing each slash character
// in path with a backslash ('\') character. Multiple separators are
// replaced by multiple backslashes.
func toBackslash(path string) string {
	return strings.Replace(path, "/", "\\", -1)
}
