package compress

import (
	tar "archive/tar"
	zip "archive/zip"
	"compress/flate"
	gzip "compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	bgzf "github.com/biogo/hts/bgzf"
	pgzip "github.com/klauspost/pgzip"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	lz4 "github.com/pierrec/lz4"
	"gopkg.in/yaml.v2"
)

type Metadata map[string]Metaitem

type Metaitem struct {
	CompSize int64  `yaml:"compsize"`
	OrigSize int64  `yaml:"origsize"`
	CompType string `yaml:"comptype"`
	CompDate string `yaml:"compdate"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	OutputPath        string `mapstructure:"output"`
	OutputFile        string `mapstructure:"file"`
	Compression       int    `mapstructure:"compression"`
	Metadata          bool   `mapstructure:"metadata"`
	NumCPU            int    `mapstructure:"numcpu"`
	Format            string `mapstructure:"format"`
	KeepInputArtifact bool   `mapstructure:"keep_input_artifact"`
	tpl               *packer.ConfigTemplate
}

type CompressPostProcessor struct {
	cfg Config
}

func (p *CompressPostProcessor) Configure(raws ...interface{}) error {
	p.cfg.Compression = -1
	_, err := common.DecodeConfig(&p.cfg, raws...)
	if err != nil {
		return err
	}

	errs := new(packer.MultiError)

	p.cfg.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.cfg.tpl.UserVars = p.cfg.PackerUserVars

	if p.cfg.OutputPath == "" {
		p.cfg.OutputPath = "packer_{{.BuildName}}_{{.Provider}}"
	}

	if err = p.cfg.tpl.Validate(p.cfg.OutputPath); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing target template: %s", err))
	}

	templates := map[string]*string{
		"output": &p.cfg.OutputPath,
	}

	if p.cfg.Compression > flate.BestCompression {
		p.cfg.Compression = flate.BestCompression
	}
	if p.cfg.Compression == -1 {
		p.cfg.Compression = flate.DefaultCompression
	}

	if p.cfg.NumCPU < 1 {
		p.cfg.NumCPU = runtime.NumCPU()
	}

	runtime.GOMAXPROCS(p.cfg.NumCPU)

	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}

		*ptr, err = p.cfg.tpl.Process(*ptr, nil)
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

func (p *CompressPostProcessor) fillMetadata(metadata Metadata, files []string) Metadata {
	// layout shows by example how the reference time should be represented.
	const layout = "2006-01-02_15-04-05"
	t := time.Now()

	if !p.cfg.Metadata {
		return metadata
	}
	for _, f := range files {
		if fi, err := os.Stat(f); err != nil {
			continue
		} else {
			if i, ok := metadata[filepath.Base(f)]; !ok {
				metadata[filepath.Base(f)] = Metaitem{CompType: p.cfg.Format, OrigSize: fi.Size(), CompDate: t.Format(layout)}
			} else {
				i.CompSize = fi.Size()
				i.CompDate = t.Format(layout)
				metadata[filepath.Base(f)] = i
			}
		}
	}
	return metadata
}

func (p *CompressPostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	newartifact := &Artifact{builderId: artifact.BuilderId(), dir: p.cfg.OutputPath}
	var metafile string = filepath.Join(p.cfg.OutputPath, "metadata")

	_, err := os.Stat(newartifact.dir)
	if err == nil {
		return nil, false, fmt.Errorf("output dir must not exists: %s", err)
	}
	err = os.MkdirAll(newartifact.dir, 0755)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create output: %s", err)
	}

	formats := strings.Split(p.cfg.Format, ".")
	files := artifact.Files()

	metadata := make(Metadata, 0)
	metadata = p.fillMetadata(metadata, files)

	for _, compress := range formats {
		switch compress {
		case "tar":
			files, err = p.cmpTAR(files, filepath.Join(p.cfg.OutputPath, p.cfg.OutputFile))
			metadata = p.fillMetadata(metadata, files)
		case "zip":
			files, err = p.cmpZIP(files, filepath.Join(p.cfg.OutputPath, p.cfg.OutputFile))
			metadata = p.fillMetadata(metadata, files)
		case "pgzip":
			files, err = p.cmpPGZIP(files, p.cfg.OutputPath)
			metadata = p.fillMetadata(metadata, files)
		case "gzip":
			files, err = p.cmpGZIP(files, p.cfg.OutputPath)
			metadata = p.fillMetadata(metadata, files)
		case "bgzf":
			files, err = p.cmpBGZF(files, p.cfg.OutputPath)
			metadata = p.fillMetadata(metadata, files)
		case "lz4":
			files, err = p.cmpLZ4(files, p.cfg.OutputPath)
			metadata = p.fillMetadata(metadata, files)
		case "e2fs":
			files, err = p.cmpE2FS(files, filepath.Join(p.cfg.OutputPath, p.cfg.OutputFile))
			metadata = p.fillMetadata(metadata, files)
		}
		if err != nil {
			return nil, false, fmt.Errorf("Failed to compress: %s", err)
		}
	}

	if p.cfg.Metadata {
		fp, err := os.Create(metafile)
		if err != nil {
			return nil, false, err
		}
		if buf, err := yaml.Marshal(metadata); err != nil {
			fp.Close()
			return nil, false, err
		} else {
			if _, err = fp.Write(buf); err != nil {
				fp.Close()
				return nil, false, err
			}
			fp.Close()
		}
	}

	newartifact.f = append(newartifact.f, files...)
	if p.cfg.Metadata {
		newartifact.f = append(newartifact.f, metafile)
	}

	return newartifact, p.cfg.KeepInputArtifact, nil
}

func (p *CompressPostProcessor) cmpTAR(src []string, dst string) ([]string, error) {
	fw, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("tar error: %s", err)
	}
	defer fw.Close()

	tw := tar.NewWriter(fw)
	defer tw.Close()

	for _, name := range src {
		fi, err := os.Stat(name)
		if err != nil {
			return nil, fmt.Errorf("tar error: %s", err)
		}

		target, _ := os.Readlink(name)
		header, err := tar.FileInfoHeader(fi, target)
		if err != nil {
			return nil, fmt.Errorf("tar erorr: %s", err)
		}

		if err = tw.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("tar error: %s", err)
		}

		fr, err := os.Open(name)
		if err != nil {
			return nil, fmt.Errorf("tar error: %s", err)
		}

		if _, err = io.Copy(tw, fr); err != nil {
			fr.Close()
			return nil, fmt.Errorf("tar error: %s", err)
		}
		fr.Close()
	}
	return []string{dst}, nil
}

func (p *CompressPostProcessor) cmpGZIP(src []string, dst string) ([]string, error) {
	var res []string
	for _, name := range src {
		filename := filepath.Join(dst, filepath.Base(name))
		fw, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("gzip error: %s", err)
		}
		cw, err := gzip.NewWriterLevel(fw, p.cfg.Compression)
		if err != nil {
			fw.Close()
			return nil, fmt.Errorf("gzip error: %s", err)
		}
		fr, err := os.Open(name)
		if err != nil {
			cw.Close()
			fw.Close()
			return nil, fmt.Errorf("gzip error: %s", err)
		}
		if _, err = io.Copy(cw, fr); err != nil {
			cw.Close()
			fr.Close()
			fw.Close()
			return nil, fmt.Errorf("gzip error: %s", err)
		}
		cw.Close()
		fr.Close()
		fw.Close()
		res = append(res, filename)
	}
	return res, nil
}

func (p *CompressPostProcessor) cmpPGZIP(src []string, dst string) ([]string, error) {
	var res []string
	for _, name := range src {
		filename := filepath.Join(dst, filepath.Base(name))
		fw, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("pgzip error: %s", err)
		}
		cw, err := pgzip.NewWriterLevel(fw, p.cfg.Compression)
		if err != nil {
			fw.Close()
			return nil, fmt.Errorf("pgzip error: %s", err)
		}
		fr, err := os.Open(name)
		if err != nil {
			cw.Close()
			fw.Close()
			return nil, fmt.Errorf("pgzip error: %s", err)
		}
		if _, err = io.Copy(cw, fr); err != nil {
			cw.Close()
			fr.Close()
			fw.Close()
			return nil, fmt.Errorf("pgzip error: %s", err)
		}
		cw.Close()
		fr.Close()
		fw.Close()
		res = append(res, filename)
	}
	return res, nil
}

func (p *CompressPostProcessor) cmpLZ4(src []string, dst string) ([]string, error) {
	var res []string
	for _, name := range src {
		filename := filepath.Join(dst, filepath.Base(name))
		fw, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("lz4 error: %s", err)
		}
		cw := lz4.NewWriter(fw)
		if err != nil {
			fw.Close()
			return nil, fmt.Errorf("lz4 error: %s", err)
		}
		if p.cfg.Compression > flate.DefaultCompression {
			cw.Header.HighCompression = true
		}
		fr, err := os.Open(name)
		if err != nil {
			cw.Close()
			fw.Close()
			return nil, fmt.Errorf("lz4 error: %s", err)
		}
		if _, err = io.Copy(cw, fr); err != nil {
			cw.Close()
			fr.Close()
			fw.Close()
			return nil, fmt.Errorf("lz4 error: %s", err)
		}
		cw.Close()
		fr.Close()
		fw.Close()
		res = append(res, filename)
	}
	return res, nil
}

func (p *CompressPostProcessor) cmpBGZF(src []string, dst string) ([]string, error) {
	var res []string
	for _, name := range src {
		filename := filepath.Join(dst, filepath.Base(name))
		fw, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("bgzf error: %s", err)
		}

		cw, err := bgzf.NewWriterLevel(fw, p.cfg.Compression, runtime.NumCPU())
		if err != nil {
			return nil, fmt.Errorf("bgzf error: %s", err)
		}
		fr, err := os.Open(name)
		if err != nil {
			cw.Close()
			fw.Close()
			return nil, fmt.Errorf("bgzf error: %s", err)
		}
		if _, err = io.Copy(cw, fr); err != nil {
			cw.Close()
			fr.Close()
			fw.Close()
			return nil, fmt.Errorf("bgzf error: %s", err)
		}
		cw.Close()
		fr.Close()
		fw.Close()
		res = append(res, filename)
	}
	return res, nil
}

func (p *CompressPostProcessor) cmpE2FS(src []string, dst string) ([]string, error) {
	panic("not implemented")
}

func (p *CompressPostProcessor) cmpZIP(src []string, dst string) ([]string, error) {
	fw, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("zip error: %s", err)
	}
	defer fw.Close()

	zw := zip.NewWriter(fw)
	defer zw.Close()

	for _, name := range src {
		header, err := zw.Create(name)
		if err != nil {
			return nil, fmt.Errorf("zip erorr: %s", err)
		}

		fr, err := os.Open(name)
		if err != nil {
			return nil, fmt.Errorf("zip error: %s", err)
		}

		if _, err = io.Copy(header, fr); err != nil {
			fr.Close()
			return nil, fmt.Errorf("zip error: %s", err)
		}
		fr.Close()
	}
	return []string{dst}, nil

}
