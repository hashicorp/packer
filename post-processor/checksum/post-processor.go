//go:generate mapstructure-to-hcl2 -type Config

package checksum

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
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

type outputPathTemplate struct {
	BuildName    string
	BuilderType  string
	ChecksumType string
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

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{"output"},
		},
	}, raws...)
	if err != nil {
		return err
	}
	errs := new(packer.MultiError)

	if p.config.ChecksumTypes == nil {
		p.config.ChecksumTypes = []string{"md5"}
	}

	for _, k := range p.config.ChecksumTypes {
		if h := getHash(k); h == nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Unrecognized checksum type: %s", k))
		}
	}

	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{.BuildName}}_{{.BuilderType}}_{{.ChecksumType}}.checksum"
	}

	if err = interpolate.Validate(p.config.OutputPath, &p.config.ctx); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing target template: %s", err))
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	files := artifact.Files()
	var h hash.Hash

	newartifact := NewArtifact(artifact.Files())
	opTpl := &outputPathTemplate{
		BuildName:   p.config.PackerBuildName,
		BuilderType: p.config.PackerBuilderType,
	}

	for _, ct := range p.config.ChecksumTypes {
		h = getHash(ct)
		opTpl.ChecksumType = ct
		p.config.ctx.Data = &opTpl

		for _, art := range files {
			checksumFile, err := interpolate.Render(p.config.OutputPath, &p.config.ctx)
			if err != nil {
				return nil, false, true, err
			}

			if _, err := os.Stat(checksumFile); err != nil {
				newartifact.files = append(newartifact.files, checksumFile)
			}
			if err := os.MkdirAll(filepath.Dir(checksumFile), os.FileMode(0755)); err != nil {
				return nil, false, true, fmt.Errorf("unable to create dir: %s", err.Error())
			}
			fw, err := os.OpenFile(checksumFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(0644))
			if err != nil {
				return nil, false, true, fmt.Errorf("unable to create file %s: %s", checksumFile, err.Error())
			}
			fr, err := os.Open(art)
			if err != nil {
				fw.Close()
				return nil, false, true, fmt.Errorf("unable to open file %s: %s", art, err.Error())
			}

			if _, err = io.Copy(h, fr); err != nil {
				fr.Close()
				fw.Close()
				return nil, false, true, fmt.Errorf("unable to compute %s hash for %s", ct, art)
			}
			fr.Close()
			fw.WriteString(fmt.Sprintf("%x\t%s\n", h.Sum(nil), filepath.Base(art)))
			fw.Close()
			h.Reset()
		}
	}

	// sets keep and forceOverride to true because we don't want to accidentally
	// delete the very artifact we're checksumming.
	return newartifact, true, true, nil
}
