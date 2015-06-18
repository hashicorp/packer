package compress

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/packer/builder/file"
	env "github.com/mitchellh/packer/helper/builder/testing"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template"
)

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

func TestSimpleCompress(t *testing.T) {
	if os.Getenv(env.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set", env.TestEnvVar))
	}

	ui, artifact, err := setup(t)
	if err != nil {
		t.Fatalf("Error bootstrapping test: %s", err)
	}
	if artifact != nil {
		defer artifact.Destroy()
	}

	tpl, err := template.Parse(strings.NewReader(simpleTestCase))
	if err != nil {
		t.Fatalf("Unable to parse test config: %s", err)
	}

	compressor := PostProcessor{}
	compressor.Configure(tpl.PostProcessors[0][0].Config)
	artifactOut, _, err := compressor.PostProcess(ui, artifact)
	if err != nil {
		t.Fatalf("Failed to compress artifact: %s", err)
	}
	// Cleanup after the test completes
	defer artifactOut.Destroy()

	// Verify things look good
	fi, err := os.Stat("package.tar.gz")
	if err != nil {
		t.Errorf("Unable to read archive: %s", err)
	}
	if fi.IsDir() {
		t.Error("Archive should not be a directory")
	}
}

func TestZipArchive(t *testing.T) {
	if os.Getenv(env.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set", env.TestEnvVar))
	}

	ui, artifact, err := setup(t)
	if err != nil {
		t.Fatalf("Error bootstrapping test: %s", err)
	}
	if artifact != nil {
		defer artifact.Destroy()
	}

	tpl, err := template.Parse(strings.NewReader(tarTestCase))
	if err != nil {
		t.Fatalf("Unable to parse test config: %s", err)
	}

	compressor := PostProcessor{}
	compressor.Configure(tpl.PostProcessors[0][0].Config)
	artifactOut, _, err := compressor.PostProcess(ui, artifact)
	if err != nil {
		t.Fatalf("Failed to archive artifact: %s", err)
	}
	// Cleanup after the test completes
	defer artifactOut.Destroy()

	// Verify things look good
	_, err = os.Stat("package.zip")
	if err != nil {
		t.Errorf("Unable to read archive: %s", err)
	}
}

func TestTarArchive(t *testing.T) {
	if os.Getenv(env.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set", env.TestEnvVar))
	}

	ui, artifact, err := setup(t)
	if err != nil {
		t.Fatalf("Error bootstrapping test: %s", err)
	}
	if artifact != nil {
		defer artifact.Destroy()
	}

	tpl, err := template.Parse(strings.NewReader(tarTestCase))
	if err != nil {
		t.Fatalf("Unable to parse test config: %s", err)
	}

	compressor := PostProcessor{}
	compressor.Configure(tpl.PostProcessors[0][0].Config)
	artifactOut, _, err := compressor.PostProcess(ui, artifact)
	if err != nil {
		t.Fatalf("Failed to archive artifact: %s", err)
	}
	// Cleanup after the test completes
	defer artifactOut.Destroy()

	// Verify things look good
	_, err = os.Stat("package.tar")
	if err != nil {
		t.Errorf("Unable to read archive: %s", err)
	}
}

func TestCompressOptions(t *testing.T) {
	if os.Getenv(env.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set", env.TestEnvVar))
	}

	ui, artifact, err := setup(t)
	if err != nil {
		t.Fatalf("Error bootstrapping test: %s", err)
	}
	if artifact != nil {
		defer artifact.Destroy()
	}

	tpl, err := template.Parse(strings.NewReader(zipTestCase))
	if err != nil {
		t.Fatalf("Unable to parse test config: %s", err)
	}

	compressor := PostProcessor{}
	compressor.Configure(tpl.PostProcessors[0][0].Config)
	artifactOut, _, err := compressor.PostProcess(ui, artifact)
	if err != nil {
		t.Fatalf("Failed to archive artifact: %s", err)
	}
	// Cleanup after the test completes
	defer artifactOut.Destroy()

	// Verify things look good
	_, err = os.Stat("package.gz")
	if err != nil {
		t.Errorf("Unable to read archive: %s", err)
	}
}

const simpleTestCase = `
{
    "post-processors": [
        {
            "type": "compress",
            "output": "package.tar.gz"
        }
    ]
}
`

const zipTestCase = `
{
    "post-processors": [
        {
            "type": "compress",
            "output": "package.zip"
        }
    ]
}
`

const tarTestCase = `
{
    "post-processors": [
        {
            "type": "compress",
            "output": "package.tar"
        }
    ]
}
`

const optionsTestCase = `
{
    "post-processors": [
        {
            "type": "compress",
            "output": "package.gz",
            "level": 9,
            "parallel": false
        }
    ]
}
`
