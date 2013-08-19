package vagrant

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
)

// OutputPathTemplate is the structure that is availalable within the
// OutputPath variables.
type OutputPathTemplate struct {
	ArtifactId string
	BuildName  string
	Provider   string
}

// Copies a file by copying the contents of the file to another place.
func CopyContents(dst, src string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	if _, err := io.Copy(dstF, srcF); err != nil {
		return err
	}

	return nil
}

// DirToBox takes the directory and compresses it into a Vagrant-compatible
// box. This function does not perform checks to verify that dir is
// actually a proper box. This is an expected precondition.
func DirToBox(dst, dir string) error {
	log.Printf("Turning dir into box: %s => %s", dir, dst)
	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	gzipWriter := gzip.NewWriter(dstF)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// This is the walk func that tars each of the files in the dir
	tarWalk := func(path string, info os.FileInfo, prevErr error) error {
		// If there was a prior error, return it
		if prevErr != nil {
			return prevErr
		}

		// Skip directories
		if info.IsDir() {
			log.Printf("Skipping directory '%s' for box '%s'", path, dst)
			return nil
		}

		log.Printf("Box add: '%s' to '%s'", path, dst)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// We have to set the Name explicitly because it is supposed to
		// be a relative path to the root. Otherwise, the tar ends up
		// being a bunch of files in the root, even if they're actually
		// nested in a dir in the original "dir" param.
		header.Name, err = filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if _, err := io.Copy(tarWriter, f); err != nil {
			return err
		}

		return nil
	}

	// Tar.gz everything up
	return filepath.Walk(dir, tarWalk)
}

// WriteMetadata writes the "metadata.json" file for a Vagrant box.
func WriteMetadata(dir string, contents interface{}) error {
	f, err := os.Create(filepath.Join(dir, "metadata.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(contents)
}
