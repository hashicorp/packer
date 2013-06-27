package vagrant

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/mitchellh/packer/packer"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

// OutputPathTemplate is the structure that is availalable within the
// OutputPath variables.
type OutputPathTemplate struct {
	ArtifactId string
	Provider   string
}

// DirToBox takes the directory and compresses it into a Vagrant-compatible
// box. This function does not perform checks to verify that dir is
// actually a proper box. This is an expected precondition.
func DirToBox(dst, dir string) error {
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

// ProcessOutputPath takes an output path template and executes it,
// replacing variables with their respective values.
func ProcessOutputPath(path string, provider string, artifact packer.Artifact) (string, error) {
	var buf bytes.Buffer

	tplData := &OutputPathTemplate{
		ArtifactId: artifact.Id(),
		Provider:   provider,
	}

	t, err := template.New("output").Parse(path)
	if err != nil {
		return "", err
	}

	err = t.Execute(&buf, tplData)
	return buf.String(), err
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
