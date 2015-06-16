package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	OutputPath string `mapstructure:"output"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (self *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&self.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	return nil

}

func (self *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	ui.Say(fmt.Sprintf("Creating archive for '%s'", artifact.BuilderId()))

	// Create the compressed archive file at the appropriate OutputPath.
	fw, err := os.Create(self.config.OutputPath)
	if err != nil {
		return nil, false, fmt.Errorf(
			"Failed creating file for compressed archive: %s", self.config.OutputPath)
	}
	defer fw.Close()

	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// Iterate through all of the artifact's files and put them into the
	// compressed archive using the tar/gzip writers.
	for _, path := range artifact.Files() {
		fi, err := os.Stat(path)
		if err != nil {
			return nil, false, fmt.Errorf(
				"Failed stating file: %s", path)
		}

		target, _ := os.Readlink(path)
		header, err := tar.FileInfoHeader(fi, target)
		if err != nil {
			return nil, false, fmt.Errorf(
				"Failed creating archive header: %s", path)
		}

		tw := tar.NewWriter(gw)
		defer tw.Close()

		// Write the header first to the archive. This takes partial data
		// from the FileInfo that is grabbed by running the stat command.
		if err := tw.WriteHeader(header); err != nil {
			return nil, false, fmt.Errorf(
				"Failed writing archive header: %s", path)
		}

		// Open the target file for archiving and compressing.
		fr, err := os.Open(path)
		if err != nil {
			return nil, false, fmt.Errorf(
				"Failed opening file '%s' to write compressed archive.", path)
		}
		defer fr.Close()

		if _, err = io.Copy(tw, fr); err != nil {
			return nil, false, fmt.Errorf(
				"Failed copying file to archive: %s", path)
		}
	}

	return NewArtifact(artifact.BuilderId(), self.config.OutputPath), false, nil
}
