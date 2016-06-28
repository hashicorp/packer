package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	Builds      []Artifact `json:"builds"`
	LastRunUUID string     `json:"last_run_uuid"`
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

	var err error

	// Create the current artifact.
	artifact.ArtifactFiles = source.Files()
	artifact.ArtifactId = source.Id()
	artifact.BuilderType = p.config.PackerBuilderType
	artifact.BuildName = p.config.PackerBuildName
	artifact.BuildTime = time.Now().Unix()
	// Since each post-processor runs in a different process we need a way to
	// coordinate between various post-processors in a single packer run. We do
	// this by setting a UUID per run and tracking this in the manifest file.
	// When we detect that the UUID in the file is the same, we know that we are
	// part of the same run and we simply add our data to the list. If the UUID
	// is different we will check the -force flag and decide whether to truncate
	// the file before we proceed.
	artifact.PackerRunUUID = os.Getenv("PACKER_RUN_UUID")

	// Create a lock file with exclusive access. If this fails we will retry
	// after a delay.
	lockFilename := p.config.Filename + ".lock"
	for i := 0; i < 3; i++ {
		// The file should not be locked for very long so we'll keep this short.
		time.Sleep((time.Duration(i) * 200 * time.Millisecond))
		_, err = os.OpenFile(lockFilename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if err == nil {
			break
		}
		log.Printf("Error locking manifest file for reading and writing. Will sleep and retry. %s", err)
	}
	defer os.Remove(lockFilename)

	// TODO fix error on first run:
	// * Post-processor failed: open packer-manifest.json: no such file or directory
	//
	// Read the current manifest file from disk
	contents := []byte{}
	if contents, err = ioutil.ReadFile(p.config.Filename); err != nil && !os.IsNotExist(err) {
		return source, true, fmt.Errorf("Unable to open %s for reading: %s", p.config.Filename, err)
	}

	// Parse the manifest file JSON, if we have one
	manifestFile := &ManifestFile{}
	if len(contents) > 0 {
		if err = json.Unmarshal(contents, manifestFile); err != nil {
			return source, true, fmt.Errorf("Unable to parse content from %s: %s", p.config.Filename, err)
		}
	}

	// If -force is set and we are not on same run, truncate the file. Otherwise
	// we will continue to add new builds to the existing manifest file.
	if p.config.PackerForce && os.Getenv("PACKER_RUN_UUID") != manifestFile.LastRunUUID {
		manifestFile = &ManifestFile{}
	}

	// Add the current artifact to the manifest file
	manifestFile.Builds = append(manifestFile.Builds, *artifact)
	manifestFile.LastRunUUID = os.Getenv("PACKER_RUN_UUID")

	// Write JSON to disk
	if out, err := json.MarshalIndent(manifestFile, "", "  "); err == nil {
		if err = ioutil.WriteFile(p.config.Filename, out, 0664); err != nil {
			return source, true, fmt.Errorf("Unable to write %s: %s", p.config.Filename, err)
		}
	} else {
		return source, true, fmt.Errorf("Unable to marshal JSON %s", err)
	}

	return source, true, nil
}
