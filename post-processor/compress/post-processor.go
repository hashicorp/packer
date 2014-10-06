package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	OutputPath string `mapstructure:"output"`

	tpl *packer.ConfigTemplate
}

type PostProcessor struct {
	config Config
}

func (self *PostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&self.config, raws...)
	if err != nil {
		return err
	}

	self.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	self.config.tpl.UserVars = self.config.PackerUserVars

	templates := map[string]*string{
		"output": &self.config.OutputPath,
	}

	errs := new(packer.MultiError)
	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}

		*ptr, err = self.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", key, err))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
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
