package compress

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/klauspost/pgzip"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"github.com/pierrec/lz4"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	OutputPath          string `mapstructure:"output"`
	Level               int    `mapstructure:"level"`
	KeepInputArtifact   bool   `mapstructure:"keep_input_artifact"`
	Archive             string
	Algorithm           string
	ctx                 *interpolate.Context
}

type PostProcessor struct {
	config Config
}

// ErrInvalidCompressionLevel is returned when the compression level passed to
// gzip is not in the expected range. See compress/flate for details.
var ErrInvalidCompressionLevel = fmt.Errorf(
	"Invalid compression level. Expected an integer from -1 to 9.")

var ErrWrongInputCount = fmt.Errorf(
	"Can only have 1 input file when not using tar/zip")

func detectFromFilename(config *Config) error {
	re := regexp.MustCompile("^.+?(?:\\.([a-z0-9]+))?\\.([a-z0-9]+)$")

	extensions := map[string]string{
		"tar": "tar",
		"zip": "zip",
		"gz":  "pgzip",
		"lz4": "lz4",
	}

	result := re.FindAllString(config.OutputPath, -1)

	// Should we make an archive? E.g. tar or zip?
	if result[0] == "tar" {
		config.Archive = "tar"
	}
	if result[1] == "zip" || result[1] == "tar" {
		config.Archive = result[1]
		// Tar or zip is our final artifact. Bail out.
		return nil
	}

	// Should we compress the artifact?
	algorithm, ok := extensions[result[1]]
	if ok {
		config.Algorithm = algorithm
		// We found our compression algorithm something. Bail out.
		return nil
	}

	// We didn't find anything. Default to tar + pgzip
	config.Algorithm = "pgzip"
	config.Archive = "tar"
	return fmt.Errorf("Unable to detect compression algorithm")
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	p.config.Level = -1
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)

	errs := new(packer.MultiError)

	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{.BuildName}}_{{.Provider}}"
	}

	if err = interpolate.Validate(p.config.OutputPath, p.config.ctx); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing target template: %s", err))
	}

	templates := map[string]*string{
		"output": &p.config.OutputPath,
	}

	if p.config.Level > gzip.BestCompression {
		p.config.Level = gzip.BestCompression
	}
	if p.config.Level == -1 {
		p.config.Level = gzip.DefaultCompression
	}

	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}

		*ptr, err = interpolate.Render(p.config.OutputPath, p.config.ctx)
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

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {

	newArtifact := &Artifact{Path: p.config.OutputPath}

	outputFile, err := os.Create(p.config.OutputPath)
	if err != nil {
		return nil, false, fmt.Errorf(
			"Unable to create archive %s: %s", p.config.OutputPath, err)
	}
	defer outputFile.Close()

	// Setup output interface. If we're using compression, output is a
	// compression writer. Otherwise it's just a file.
	var output io.WriteCloser
	switch p.config.Algorithm {
	case "lz4":
		lzwriter := lz4.NewWriter(outputFile)
		if p.config.Level > gzip.DefaultCompression {
			lzwriter.Header.HighCompression = true
		}
		defer lzwriter.Close()
		output = lzwriter
	case "pgzip":
		output, err = pgzip.NewWriterLevel(outputFile, p.config.Level)
		if err != nil {
			return nil, false, ErrInvalidCompressionLevel
		}
		defer output.Close()
	default:
		output = outputFile
	}

	//Archive
	switch p.config.Archive {
	case "tar":
		archiveTar(artifact.Files(), output)
	case "zip":
		archive := zip.NewWriter(output)
		defer archive.Close()
	default:
		// We have a regular file, so we'll just do an io.Copy
		if len(artifact.Files()) != 1 {
			return nil, false, fmt.Errorf(
				"Can only have 1 input file when not using tar/zip. Found %d "+
					"files: %v", len(artifact.Files()), artifact.Files())
		}
		source, err := os.Open(artifact.Files()[0])
		if err != nil {
			return nil, false, fmt.Errorf(
				"Failed to open source file %s for reading: %s",
				artifact.Files()[0], err)
		}
		defer source.Close()
		io.Copy(output, source)
	}

	return newArtifact, p.config.KeepInputArtifact, nil
}

func archiveTar(files []string, output io.WriteCloser) error {
	archive := tar.NewWriter(output)
	defer archive.Close()

	for _, path := range files {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Unable to read file %s: %s", path, err)
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			return fmt.Errorf("Unable to get fileinfo for %s: %s", path, err)
		}

		target, err := os.Readlink(path)
		if err != nil {
			return fmt.Errorf("Failed to readlink for %s: %s", path, err)
		}

		header, err := tar.FileInfoHeader(fi, target)
		if err != nil {
			return fmt.Errorf("Failed to create tar header for %s: %s", path, err)
		}

		if err := archive.WriteHeader(header); err != nil {
			return fmt.Errorf("Failed to write tar header for %s: %s", path, err)
		}

		if _, err := io.Copy(archive, file); err != nil {
			return fmt.Errorf("Failed to copy %s data to archive: %s", path, err)
		}
	}
	return nil
}

func (p *PostProcessor) cmpTAR(files []string, target string) ([]string, error) {
	fw, err := os.Create(target)
	if err != nil {
		return nil, fmt.Errorf("tar error creating tar %s: %s", target, err)
	}
	defer fw.Close()

	tw := tar.NewWriter(fw)
	defer tw.Close()

	for _, name := range files {
		fi, err := os.Stat(name)
		if err != nil {
			return nil, fmt.Errorf("tar error on stat of %s: %s", name, err)
		}

		target, _ := os.Readlink(name)
		header, err := tar.FileInfoHeader(fi, target)
		if err != nil {
			return nil, fmt.Errorf("tar error reading info for %s: %s", name, err)
		}

		if err = tw.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("tar error writing header for %s: %s", name, err)
		}

		fr, err := os.Open(name)
		if err != nil {
			return nil, fmt.Errorf("tar error opening file %s: %s", name, err)
		}

		if _, err = io.Copy(tw, fr); err != nil {
			fr.Close()
			return nil, fmt.Errorf("tar error copying contents of %s: %s", name, err)
		}
		fr.Close()
	}
	return []string{target}, nil
}

func (p *PostProcessor) cmpGZIP(files []string, target string) ([]string, error) {
	var res []string
	for _, name := range files {
		filename := filepath.Join(target, filepath.Base(name))
		fw, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("gzip error creating archive: %s", err)
		}
		cw, err := gzip.NewWriterLevel(fw, p.config.Level)
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

func (p *PostProcessor) cmpPGZIP(files []string, target string) ([]string, error) {
	var res []string
	for _, name := range files {
		filename := filepath.Join(target, filepath.Base(name))
		fw, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("pgzip error: %s", err)
		}
		cw, err := pgzip.NewWriterLevel(fw, p.config.Level)
		cw.SetConcurrency(500000, runtime.GOMAXPROCS(-1))
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

func (p *PostProcessor) cmpLZ4(src []string, dst string) ([]string, error) {
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
		if p.config.Level > gzip.DefaultCompression {
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

func (p *PostProcessor) cmpZIP(src []string, dst string) ([]string, error) {
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
			return nil, fmt.Errorf("zip error: %s", err)
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
