package checksum

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/packer/builder/file"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template"
)

const expectedFileContents = "Hello world!"

func TestChecksumSHA1(t *testing.T) {
	const config = `
	{
	    "post-processors": [
	        {
	            "type": "checksum",
	            "checksum_types": ["sha1"],
	            "output": "sha1sums"
	        }
	    ]
	}
	`
	artifact := testChecksum(t, config)
	defer artifact.Destroy()

	f, err := os.Open("sha1sums")
	if err != nil {
		t.Errorf("Unable to read checksum file: %s", err)
	}
	if buf, _ := ioutil.ReadAll(f); !bytes.Equal(buf, []byte("d3486ae9136e7856bc42212385ea797094475802\tpackage.txt\n")) {
		t.Errorf("Failed to compate checksum: %s\n%s", buf, "d3486ae9136e7856bc42212385ea797094475802 package.txt")
	}

	defer f.Close()
}

// Test Helpers

func setup(t *testing.T) (packer.Ui, packer.Artifact, error) {
	// Create fake UI and Cache
	ui := packer.TestUi(t)
	cache := &packer.FileCache{CacheDir: os.TempDir()}

	// Create config for file builder
	const fileConfig = `{"builders":[{"type":"file","target":"package.txt","content":"Hello world!"}]}`
	tpl, err := template.Parse(strings.NewReader(fileConfig))
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to parse setup configuration: %s", err)
	}

	// Prepare the file builder
	builder := file.Builder{}
	warnings, err := builder.Prepare(tpl.Builders["file"].Config)
	if len(warnings) > 0 {
		for _, warn := range warnings {
			return nil, nil, fmt.Errorf("Configuration warning: %s", warn)
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("Invalid configuration: %s", err)
	}

	// Run the file builder
	artifact, err := builder.Run(ui, nil, cache)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to build artifact: %s", err)
	}

	return ui, artifact, err
}

func testChecksum(t *testing.T, config string) packer.Artifact {
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

	checksum := PostProcessor{}
	checksum.Configure(tpl.PostProcessors[0][0].Config)

	// I get the feeling these should be automatically available somewhere, but
	// some of the post-processors construct this manually.
	checksum.config.ctx.BuildName = "chocolate"
	checksum.config.PackerBuildName = "vanilla"
	checksum.config.PackerBuilderType = "file"

	artifactOut, _, err := checksum.PostProcess(ui, artifact)
	if err != nil {
		t.Fatalf("Failed to checksum artifact: %s", err)
	}

	return artifactOut
}
