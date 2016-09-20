package checksum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Keep          bool     `mapstructure:"keep_input_artifact"`
	ChecksumTypes []string `mapstructure:"checksum_types"`
	OutputPath    string   `mapstructure:"output"`
	ctx           interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.ChecksumTypes == nil {
		p.config.ChecksumTypes = []string{"md5"}
	}

	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{.BuildName}}_{{.BuilderType}}" + ".checksum"
	}

	errs := new(packer.MultiError)

	if err = interpolate.Validate(p.config.OutputPath, &p.config.ctx); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing target template: %s", err))
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func getHash(t string) hash.Hash {
	var h hash.Hash
	switch t {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha224":
		h = sha256.New224()
	case "sha256":
		h = sha256.New()
	case "sha384":
		h = sha512.New384()
	case "sha512":
		h = sha512.New()
	}
	return h
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	files := artifact.Files()
	var h hash.Hash
	var checksumFile string

	newartifact := NewArtifact(artifact.Files())

	for _, ct := range p.config.ChecksumTypes {
		h = getHash(ct)

		for _, art := range files {
			if len(artifact.Files()) > 1 {
				checksumFile = filepath.Join(filepath.Dir(art), ct+"sums")
			} else if p.config.OutputPath != "" {
				checksumFile = p.config.OutputPath
			} else {
				checksumFile = fmt.Sprintf("%s.%s", art, ct+"sum")
			}
			if _, err := os.Stat(checksumFile); err != nil {
				newartifact.files = append(newartifact.files, checksumFile)
			}
			if err := os.MkdirAll(filepath.Dir(checksumFile), os.FileMode(0755)); err != nil {
				return nil, false, fmt.Errorf("unable to create dir: %s", err.Error())
			}
			fw, err := os.OpenFile(checksumFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(0644))
			if err != nil {
				return nil, false, fmt.Errorf("unable to create file %s: %s", checksumFile, err.Error())
			}
			fr, err := os.Open(art)
			if err != nil {
				fw.Close()
				return nil, false, fmt.Errorf("unable to open file %s: %s", art, err.Error())
			}

			if _, err = io.Copy(h, fr); err != nil {
				fr.Close()
				fw.Close()
				return nil, false, fmt.Errorf("unable to compute %s hash for %s", ct, art)
			}
			fr.Close()
			fw.WriteString(fmt.Sprintf("%x\t%s\n", h.Sum(nil), filepath.Base(art)))
			fw.Close()
		}
	}

	return newartifact, true, nil
}
