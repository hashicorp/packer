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

var (
	// ErrInvalidCompressionLevel is returned when the compression level passed
	// to gzip is not in the expected range. See compress/flate for details.
	ErrInvalidCompressionLevel = fmt.Errorf(
		"Invalid compression level. Expected an integer from -1 to 9.")

	ErrWrongInputCount = fmt.Errorf(
		"Can only have 1 input file when not using tar/zip")

	filenamePattern = regexp.MustCompile(`(?:\.([a-z0-9]+))`)
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Fields from config file
	OutputPath        string `mapstructure:"output"`
	CompressionLevel  int    `mapstructure:"compression_level"`
	KeepInputArtifact bool   `mapstructure:"keep_input_artifact"`

	// Derived fields
	Archive   string
	Algorithm string

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)

	errs := new(packer.MultiError)

	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{.BuildName}}_{{.Provider}}"
	}

	if err = interpolate.Validate(p.config.OutputPath, &p.config.ctx); err != nil {
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

	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}

		*ptr, err = interpolate.Render(p.config.OutputPath, &p.config.ctx)
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
	keep := p.config.KeepInputArtifact
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
		ui.Say(fmt.Sprintf("Using lz4 compression with %d cores for %s",
			runtime.GOMAXPROCS(-1), target))
		output, err = makeLZ4Writer(outputFile, p.config.CompressionLevel)
		defer output.Close()
	case "pgzip":
		ui.Say(fmt.Sprintf("Using pgzip compression with %d cores for %s",
			runtime.GOMAXPROCS(-1), target))
		output, err = makePgzipWriter(outputFile, p.config.CompressionLevel)
		defer output.Close()
	default:
		output = outputFile
	}

	compression := p.config.Algorithm
	if compression == "" {
		compression = "no compression"
	}

	// Build an archive, if we're supposed to do that.
	switch p.config.Archive {
	case "tar":
		ui.Say(fmt.Sprintf("Tarring %s with %s", target, compression))
		err = createTarArchive(artifact.Files(), output)
		if err != nil {
			return nil, keep, fmt.Errorf("Error creating tar: %s", err)
		}
	case "zip":
		ui.Say(fmt.Sprintf("Zipping %s", target))
		err = createZipArchive(artifact.Files(), output)
		if err != nil {
			return nil, keep, fmt.Errorf("Error creating zip: %s", err)
		}
	default:
		// Filename indicates no tarball (just compress) so we'll do an io.Copy
		// into our compressor.
		if len(artifact.Files()) != 1 {
			return nil, keep, fmt.Errorf(
				"Can only have 1 input file when not using tar/zip. Found %d "+
					"files: %v", len(artifact.Files()), artifact.Files())
		}
		archiveFile := artifact.Files()[0]
		ui.Say(fmt.Sprintf("Archiving %s with %s", archiveFile, compression))

		source, err := os.Open(archiveFile)
		if err != nil {
			return nil, keep, fmt.Errorf(
				"Failed to open source file %s for reading: %s",
				archiveFile, err)
		}
		defer source.Close()

		if _, err = io.Copy(output, source); err != nil {
			return nil, keep, fmt.Errorf("Failed to compress %s: %s",
				archiveFile, err)
		}
	}

	ui.Say(fmt.Sprintf("Archive %s completed", target))

	return newArtifact, keep, nil
}

func (config *Config) detectFromFilename() {

	extensions := map[string]string{
		"tar": "tar",
		"zip": "zip",
		"gz":  "pgzip",
		"lz4": "lz4",
	}

	result := filenamePattern.FindAllStringSubmatch(config.OutputPath, -1)

	// No dots. Bail out with defaults.
	if len(result) == 0 {
		config.Algorithm = "pgzip"
		config.Archive = "tar"
		return
	}

	// Parse the last two .groups, if they're there
	lastItem := result[len(result)-1][1]
	var nextToLastItem string
	if len(result) == 1 {
		nextToLastItem = ""
	} else {
		nextToLastItem = result[len(result)-2][1]
	}

	// Should we make an archive? E.g. tar or zip?
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

	// We didn't match a known compression format. Default to tar + pgzip
	config.Algorithm = "pgzip"
	config.Archive = "tar"
	return
}

func makeLZ4Writer(output io.WriteCloser, compressionLevel int) (io.WriteCloser, error) {
	lzwriter := lz4.NewWriter(output)
	if compressionLevel > gzip.DefaultCompression {
		lzwriter.Header.HighCompression = true
	}
	return lzwriter, nil
}

func makePgzipWriter(output io.WriteCloser, compressionLevel int) (io.WriteCloser, error) {
	gzipWriter, err := pgzip.NewWriterLevel(output, compressionLevel)
	if err != nil {
		return nil, ErrInvalidCompressionLevel
	}
	gzipWriter.SetConcurrency(500000, runtime.GOMAXPROCS(-1))
	return gzipWriter, nil
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

		header, err := tar.FileInfoHeader(fi, path)
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
	archive := zip.NewWriter(output)
	defer archive.Close()

	for _, path := range files {
		path = filepath.ToSlash(path)

		source, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Unable to read file %s: %s", path, err)
		}
		defer source.Close()

		target, err := archive.Create(path)
		if err != nil {
			return fmt.Errorf("Failed to add zip header for %s: %s", path, err)
		}

		_, err = io.Copy(target, source)
		if err != nil {
			return fmt.Errorf("Failed to copy %s data to archive: %s", path, err)
		}
	}
	return nil
}
