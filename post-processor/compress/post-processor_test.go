// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package compress

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dsnet/compress/bzip2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template"
	"github.com/hashicorp/packer/builder/file"
	"github.com/pierrec/lz4/v4"
)

func TestDetectFilename(t *testing.T) {
	// Test default / fallback with no file extension
	nakedFilename := Config{OutputPath: "test"}
	nakedFilename.detectFromFilename()
	if nakedFilename.Archive != "tar" {
		t.Error("Expected to find tar archive setting")
	}
	if nakedFilename.Algorithm != "pgzip" {
		t.Error("Expected to find pgzip algorithm setting")
	}

	// Test .archive
	zipFilename := Config{OutputPath: "test.zip"}
	zipFilename.detectFromFilename()
	if zipFilename.Archive != "zip" {
		t.Error("Expected to find zip archive setting")
	}
	if zipFilename.Algorithm != "" {
		t.Error("Expected to find empty algorithm setting")
	}

	// Test .compress
	lz4Filename := Config{OutputPath: "test.lz4"}
	lz4Filename.detectFromFilename()
	if lz4Filename.Archive != "" {
		t.Error("Expected to find empty archive setting")
	}
	if lz4Filename.Algorithm != "lz4" {
		t.Error("Expected to find lz4 algorithm setting")
	}

	// Test .archive.compress with some.extra.dots...
	lotsOfDots := Config{OutputPath: "test.blah.bloo.blee.tar.lz4"}
	lotsOfDots.detectFromFilename()
	if lotsOfDots.Archive != "tar" {
		t.Error("Expected to find tar archive setting")
	}
	if lotsOfDots.Algorithm != "lz4" {
		t.Error("Expected to find lz4 algorithm setting")
	}
}

const expectedFileContents = "Hello world!"

func TestCompressOptions(t *testing.T) {
	const config = `
	{
	    "post-processors": [
	        {
	            "type": "compress",
	            "output": "package.gz",
	            "compression_level": 9
	        }
	    ]
	}
	`

	artifact := testArchive(t, config)
	defer artifact.Destroy()

	filename := "package.gz"
	archive, _ := os.Open(filename)
	gzipReader, _ := gzip.NewReader(archive)
	data, _ := io.ReadAll(gzipReader)

	if string(data) != expectedFileContents {
		t.Errorf("Expected:\n%s\nFound:\n%s\n", expectedFileContents, data)
	}
}

func TestCompressInterpolation(t *testing.T) {
	const config = `
	{
	    "post-processors": [
	        {
	            "type": "compress",
	            "output": "{{ build_name}}-{{ .BuildName }}-{{.BuilderType}}.gz"
	        }
	    ]
	}
	`

	artifact := testArchive(t, config)
	defer artifact.Destroy()

	// You can interpolate using the .BuildName variable or build_name global
	// function. We'll check both.
	filename := "chocolate-vanilla-file.gz"
	archive, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Unable to read %s: %s", filename, err)
	}

	gzipReader, _ := gzip.NewReader(archive)
	data, _ := io.ReadAll(gzipReader)

	if string(data) != expectedFileContents {
		t.Errorf("Expected:\n%s\nFound:\n%s\n", expectedFileContents, data)
	}
}

// Test Helpers

func setup(t *testing.T) (packersdk.Ui, packersdk.Artifact, error) {
	// Create fake UI and Cache
	ui := packersdk.TestUi(t)

	// Create config for file builder
	const fileConfig = `{"builders":[{"type":"file","target":"package.txt","content":"Hello world!"}]}`
	tpl, err := template.Parse(strings.NewReader(fileConfig))
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to parse setup configuration: %s", err)
	}

	// Prepare the file builder
	builder := file.Builder{}
	_, warnings, err := builder.Prepare(tpl.Builders["file"].Config)
	if len(warnings) > 0 {
		for _, warn := range warnings {
			return nil, nil, fmt.Errorf("Configuration warning: %s", warn)
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("Invalid configuration: %s", err)
	}

	// Run the file builder
	artifact, err := builder.Run(context.Background(), ui, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to build artifact: %s", err)
	}

	return ui, artifact, err
}

func testArchive(t *testing.T, config string) packersdk.Artifact {
	ui, artifact, err := setup(t)
	if err != nil {
		t.Fatalf("Error bootstrapping test: %s", err)
	}
	if artifact != nil {
		defer artifact.Destroy()
	}

	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("Unable to parse test config: %s", err)
	}

	compressor := PostProcessor{}
	compressor.Configure(tpl.PostProcessors[0][0].Config)

	// I get the feeling these should be automatically available somewhere, but
	// some of the post-processors construct this manually.
	compressor.config.ctx.BuildName = "chocolate"
	compressor.config.PackerBuildName = "vanilla"
	compressor.config.PackerBuilderType = "file"

	artifactOut, _, _, err := compressor.PostProcess(context.Background(), ui, artifact)
	if err != nil {
		t.Fatalf("Failed to compress artifact: %s", err)
	}

	return artifactOut
}

func TestArchive(t *testing.T) {
	tc := map[string]func(*os.File) ([]byte, error){
		"bzip2": func(archive *os.File) ([]byte, error) {
			bzipReader, err := bzip2.NewReader(archive, nil)
			if err != nil {
				return nil, err
			}
			return io.ReadAll(bzipReader)
		},
		"zip": func(archive *os.File) ([]byte, error) {
			fi, _ := archive.Stat()
			zipReader, err := zip.NewReader(archive, fi.Size())
			if err != nil {
				return nil, err
			}
			ctt, err := zipReader.File[0].Open()
			if err != nil {
				return nil, err
			}
			return io.ReadAll(ctt)
		},
		"tar": func(archive *os.File) ([]byte, error) {
			tarReader := tar.NewReader(archive)
			_, err := tarReader.Next()
			if err != nil {
				return nil, err
			}
			return io.ReadAll(tarReader)
		},
		"tar.gz": func(archive *os.File) ([]byte, error) {
			gzipReader, err := gzip.NewReader(archive)
			if err != nil {
				return nil, err
			}
			tarReader := tar.NewReader(gzipReader)
			_, err = tarReader.Next()
			if err != nil {
				return nil, err
			}
			return io.ReadAll(tarReader)
		},
		"gz": func(archive *os.File) ([]byte, error) {
			gzipReader, _ := gzip.NewReader(archive)
			return io.ReadAll(gzipReader)
		},
		"lz4": func(archive *os.File) ([]byte, error) {
			lz4Reader := lz4.NewReader(archive)
			return io.ReadAll(lz4Reader)
		},
	}

	tmpArchiveFile := "temp-archive-package"
	for format, unzip := range tc {
		t.Run(format, func(t *testing.T) {
			config := fmt.Sprintf(`
		{
			"post-processors": [
				{
					"type": "compress",
					"output": "%s.%s"
				}
			]
		}
		`, tmpArchiveFile, format)

			artifact := testArchive(t, config)
			defer func() {
				err := artifact.Destroy()
				if err != nil {
					t.Fatal(err)
				}
			}()

			filename := fmt.Sprintf("%s.%s", tmpArchiveFile, format)
			// Verify things look good
			_, err := os.Stat(filename)
			if err != nil {
				t.Errorf("Unable to read archive: %s", err)
			}

			archive, err := os.Open(filename)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				err := archive.Close()
				if err != nil {
					t.Fatal(err)
				}
			}()

			found, err := unzip(archive)
			if err != nil {
				t.Fatal(err)
			}
			if string(found) != expectedFileContents {
				t.Errorf("Expected:\n%s\nFound:\n%s\n", expectedFileContents, string(found))
			}
		})
	}

}
