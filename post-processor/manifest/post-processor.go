package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Filename string `mapstructure:"filename"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

type ManifestFile struct {
	Builds []Artifact `json:"builds"`
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.Filename == "" {
		p.config.Filename = "packer-manifest.json"
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, source packer.Artifact) (packer.Artifact, bool, error) {
	artifact := &Artifact{}

	// Create the current artifact.
	artifact.ArtifactFiles = source.Files()
	artifact.ArtifactId = source.Id()
	artifact.BuilderType = p.config.PackerBuilderType
	artifact.BuildName = p.config.PackerBuildName
	artifact.BuildTime = time.Now().Unix()

	// Create a lock file with exclusive access. If this fails we will retry
	// after a delay
	// TODO add retry
	lockFilename := p.config.Filename + ".lock"
	_, err := os.OpenFile(lockFilename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	defer os.Remove(lockFilename)

	// TODO fix error on first run:
	// * Post-processor failed: open packer-manifest.json: no such file or directory
	//
	// Read the current manifest file from disk
	contents := []byte{}
	if contents, err = ioutil.ReadFile(p.config.Filename); err != nil && !os.IsNotExist(err) {
		return source, true, fmt.Errorf("Unable to open %s for reading: %s", p.config.Filename, err)
	}

	// Parse the manifest file JSON, if we have some
	manifestFile := &ManifestFile{}
	if len(contents) > 0 {
		if err = json.Unmarshal(contents, manifestFile); err != nil {
			return source, true, fmt.Errorf("Unable to parse content from %s: %s", p.config.Filename, err)
		}
	}

	// Add the current artifact to the manifest file
	manifestFile.Builds = append(manifestFile.Builds, *artifact)

	// Write JSON to disk
	if out, err := json.MarshalIndent(manifestFile, "", "  "); err == nil {
		if err := ioutil.WriteFile(p.config.Filename, out, 0664); err != nil {
			return source, true, fmt.Errorf("Unable to write %s: %s", p.config.Filename, err)
		}
	} else {
		return source, true, fmt.Errorf("Unable to marshal JSON %s", err)
	}

	return source, true, err
}
