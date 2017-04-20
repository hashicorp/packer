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
	BuildName   string
	BuilderType string
	ArtifactID  string
	HashType    string
}

func getHashMap() map[string]func() hash.Hash {
	return map[string]func() hash.Hash{
		"md5":    func() hash.Hash { return md5.New() },
		"sha1":   func() hash.Hash { return sha1.New() },
		"sha224": func() hash.Hash { return sha256.New224() },
		"sha256": func() hash.Hash { return sha256.New() },
		"sha384": func() hash.Hash { return sha512.New384() },
		"sha512": func() hash.Hash { return sha512.New() },
	}
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
	errs := new(packer.MultiError)

	if p.config.ChecksumTypes == nil {
		p.config.ChecksumTypes = []string{"md5"}
	}

	hashMap := getHashMap()
	for _, k := range p.config.ChecksumTypes {
		if _, ok := hashMap[k]; !ok {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Unrecognized checksum type: %s", k))
		}
	}

	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{.BuildName}}_{{.BuilderType}}_{{.HashType}}.checksum"
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

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	files := artifact.Files()
	var h hash.Hash

	newartifact := NewArtifact(artifact.Files())
	opTpl := &outputPathTemplate{
		BuildName:   p.config.PackerBuildName,
		BuilderType: p.config.PackerBuilderType,
		ArtifactID:  artifact.Id(),
	}

	for _, ct := range p.config.ChecksumTypes {
		h = getHashMap()[ct]()
		opTpl.HashType = ct
		p.config.ctx.Data = &opTpl

		for _, art := range files {
			checksumFile, err := interpolate.Render(p.config.OutputPath, &p.config.ctx)
			if err != nil {
				return nil, false, err
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
			h.Reset()
		}
	}

	return newartifact, true, nil
}
