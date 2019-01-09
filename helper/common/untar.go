package common

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
)

// UntarBox takes a vagrant .box file and decompresses it into the target
// directory.
func UntarBox(dir, src string) error {
	log.Printf("Turning .box to dir: %s => %s", src, dir)
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	gzr, err := gzip.NewReader(srcF)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)
	for {
		hdr, err := tarReader.Next()
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF {
			break
		}

		// We use the fileinfo to get the file name because we are not
		// expecting path information as from the tar header. It's important
		// that we not use the path name from the tar header without checking
		// for the presence of `..`. If we accidentally allow for that, we can
		// open ourselves up to a path traversal vulnerability.
		info := hdr.FileInfo()

		// Shouldn't be any directories, skip them
		if info.IsDir() {
			continue
		}

		// We wrap this in an anonymous function so that the defers
		// inside are handled more quickly so we can give up file handles.
		err = func() error {
			path := filepath.Join(dir, info.Name())
			output, err := os.Create(path)
			if err != nil {
				return err
			}
			defer output.Close()

			os.Chmod(path, info.Mode())
			os.Chtimes(path, hdr.AccessTime, hdr.ModTime)
			_, err = io.Copy(output, tarReader)
			return err
		}()
		if err != nil {
			return err
		}
	}

	return nil
}
