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
	CompressionLevel    int    `mapstructure:"compression_level"`
	KeepInputArtifact   bool   `mapstructure:"keep_input_artifact"`
	Archive             string
	Algorithm           string
	UsingDefault        bool
	ctx                 *interpolate.Context
}

type PostProcessor struct {
	config *Config
}

var (
	// ErrInvalidCompressionLevel is returned when the compression level passed
	// to gzip is not in the expected range. See compress/flate for details.
	ErrInvalidCompressionLevel = fmt.Errorf(
		"Invalid compression level. Expected an integer from -1 to 9.")

	ErrWrongInputCount = fmt.Errorf(
		"Can only have 1 input file when not using tar/zip")

	filenamePattern = regexp.MustCompile(`(?:\.([a-z0-9]+))`)
)

func (config *Config) detectFromFilename() {

	extensions := map[string]string{
		"tar": "tar",
		"zip": "zip",
		"gz":  "pgzip",
		"lz4": "lz4",
	}

	result := filenamePattern.FindAllStringSubmatch(config.OutputPath, -1)

	if len(result) == 0 {
		config.Algorithm = "pgzip"
		config.Archive = "tar"
		return
	}

	// Should we make an archive? E.g. tar or zip?
	var nextToLastItem string
	if len(result) == 1 {
		nextToLastItem = ""
	} else {
		nextToLastItem = result[len(result)-2][1]
	}

	lastItem := result[len(result)-1][1]
	if nextToLastItem == "tar" {
		config.Archive = "tar"
	}
	if lastItem == "zip" || lastItem == "tar" {
		config.Archive = lastItem
		// Tar or zip is our final artifact. Bail out.
		return
	}

	// Should we compress the artifact?
	algorithm, ok := extensions[lastItem]
	if ok {
		config.Algorithm = algorithm
		// We found our compression algorithm. Bail out.
		return
	}

	// We didn't find anything. Default to tar + pgzip
	config.Algorithm = "pgzip"
	config.Archive = "tar"
	return
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)

	fmt.Printf("CompressionLevel: %d\n", p.config.CompressionLevel)

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

	if p.config.CompressionLevel > pgzip.BestCompression {
		p.config.CompressionLevel = pgzip.BestCompression
	}
	// Technically 0 means "don't compress" but I don't know how to
	// differentiate between "user entered zero" and "user entered nothing".
	// Also, why bother creating a compressed file with zero compression?
	if p.config.CompressionLevel == -1 || p.config.CompressionLevel == 0 {
		p.config.CompressionLevel = pgzip.DefaultCompression
	}

	fmt.Printf("CompressionLevel: %d\n", p.config.CompressionLevel)
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

	p.config.detectFromFilename()

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil

}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {

	target := p.config.OutputPath
	newArtifact := &Artifact{Path: target}

	outputFile, err := os.Create(target)
	if err != nil {
		return nil, false, fmt.Errorf(
			"Unable to create archive %s: %s", target, err)
	}
	defer outputFile.Close()

	// Setup output interface. If we're using compression, output is a
	// compression writer. Otherwise it's just a file.
	var output io.WriteCloser
	switch p.config.Algorithm {
	case "lz4":
		ui.Say(fmt.Sprintf("Preparing lz4 compression for %s", target))
		lzwriter := lz4.NewWriter(outputFile)
		if p.config.CompressionLevel > gzip.DefaultCompression {
			lzwriter.Header.HighCompression = true
		}
		defer lzwriter.Close()
		output = lzwriter
	case "pgzip":
		ui.Say(fmt.Sprintf("Preparing gzip compression for %s", target))
		gzipWriter, err := pgzip.NewWriterLevel(outputFile, p.config.CompressionLevel)
		if err != nil {
			return nil, false, ErrInvalidCompressionLevel
		}
		gzipWriter.SetConcurrency(500000, runtime.GOMAXPROCS(-1))
		output = gzipWriter
		defer output.Close()
	default:
		output = outputFile
	}

	compression := p.config.Algorithm
	if compression == "" {
		compression = "no"
	}

	// Build an archive, if we're supposed to do that.
	switch p.config.Archive {
	case "tar":
		ui.Say(fmt.Sprintf("Taring %s with %s compression", target, compression))
		createTarArchive(artifact.Files(), output)
	case "zip":
		ui.Say(fmt.Sprintf("Zipping %s", target))
		archive := zip.NewWriter(output)
		defer archive.Close()
	default:
		ui.Say(fmt.Sprintf("Copying %s with %s compression", target, compression))
		// Filename indicates no tarball (just compress) so we'll do an io.Copy
		// into our compressor.
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

	ui.Say(fmt.Sprintf("Archive %s completed", target))

	return newArtifact, p.config.KeepInputArtifact, nil
}

func createTarArchive(files []string, output io.WriteCloser) error {
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

func createZipArchive(files []string, output io.WriteCloser) error {
	return fmt.Errorf("Not implemented")
}

func (p *PostProcessor) cmpGZIP(files []string, target string) ([]string, error) {
	var res []string
	for _, name := range files {
		filename := filepath.Join(target, filepath.Base(name))
		fw, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("gzip error creating archive: %s", err)
		}
		cw, err := gzip.NewWriterLevel(fw, p.config.CompressionLevel)
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
		cw, err := pgzip.NewWriterLevel(fw, p.config.CompressionLevel)

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
		if p.config.CompressionLevel > gzip.DefaultCompression {
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
