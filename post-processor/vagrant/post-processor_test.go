package vagrant

import (
	"bytes"
	"compress/flate"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func testPP(t *testing.T) *PostProcessor {
	var p PostProcessor
	if err := p.Configure(testConfig()); err != nil {
		t.Fatalf("err: %s", err)
	}

	return &p
}

func testUi() *packer.BasicUi {
	return &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}
}

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packer.PostProcessor = new(PostProcessor)
}

func TestPostProcessorPrepare_compressionLevel(t *testing.T) {
	var p PostProcessor

	// Default
	c := testConfig()
	delete(c, "compression_level")
	if err := p.Configure(c); err != nil {
		t.Fatalf("err: %s", err)
	}

	config := p.config
	if config.CompressionLevel != flate.DefaultCompression {
		t.Fatalf("bad: %#v", config.CompressionLevel)
	}

	// Set
	c = testConfig()
	c["compression_level"] = 7
	if err := p.Configure(c); err != nil {
		t.Fatalf("err: %s", err)
	}

	config = p.config
	if config.CompressionLevel != 7 {
		t.Fatalf("bad: %#v", config.CompressionLevel)
	}
}

func TestPostProcessorPrepare_outputPath(t *testing.T) {
	var p PostProcessor

	// Default
	c := testConfig()
	delete(c, "output")
	err := p.Configure(c)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Bad template
	c["output"] = "bad {{{{.Template}}}}"
	err = p.Configure(c)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestSpecificConfig(t *testing.T) {
	var p PostProcessor

	// Default
	c := testConfig()
	c["compression_level"] = 1
	c["output"] = "folder"
	c["override"] = map[string]interface{}{
		"aws": map[string]interface{}{
			"compression_level": 7,
		},
	}
	if err := p.Configure(c); err != nil {
		t.Fatalf("err: %s", err)
	}

	// overrides config
	config, err := p.specificConfig("aws")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if config.CompressionLevel != 7 {
		t.Fatalf("bad: %#v", config.CompressionLevel)
	}

	if config.OutputPath != "folder" {
		t.Fatalf("bad: %#v", config.OutputPath)
	}

	// does NOT overrides config
	config, err = p.specificConfig("virtualbox")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if config.CompressionLevel != 1 {
		t.Fatalf("bad: %#v", config.CompressionLevel)
	}

	if config.OutputPath != "folder" {
		t.Fatalf("bad: %#v", config.OutputPath)
	}
}

func TestPostProcessorPrepare_vagrantfileTemplateExists(t *testing.T) {
	f, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	name := f.Name()
	c := testConfig()
	c["vagrantfile_template"] = name

	if err := f.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}

	var p PostProcessor

	if err := p.Configure(c); err != nil {
		t.Fatal("no error expected as vagrantfile_template exists")
	}

	if err := os.Remove(name); err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := p.Configure(c); err == nil {
		t.Fatal("expected error since vagrantfile_template does not exist and vagrantfile_template_generated is unset")
	}

	// The vagrantfile_template will be generated during the build process
	c["vagrantfile_template_generated"] = true

	if err := p.Configure(c); err != nil {
		t.Fatal("no error expected due to missing vagrantfile_template as vagrantfile_template_generated is set")
	}
}

func TestPostProcessorPostProcess_badId(t *testing.T) {
	artifact := &packer.MockArtifact{
		BuilderIdValue: "invalid.packer",
	}

	_, _, _, err := testPP(t).PostProcess(context.Background(), testUi(), artifact)
	if !strings.Contains(err.Error(), "artifact type") {
		t.Fatalf("err: %s", err)
	}
}

func TestPostProcessorPostProcess_vagrantfileUserVariable(t *testing.T) {
	var p PostProcessor

	f, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(f.Name())

	c := map[string]interface{}{
		"packer_user_variables": map[string]string{
			"foo": f.Name(),
		},

		"vagrantfile_template": "{{user `foo`}}",
	}
	err = p.Configure(c)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	a := &packer.MockArtifact{
		BuilderIdValue: "packer.parallels",
	}
	a2, _, _, err := p.PostProcess(context.Background(), testUi(), a)
	if a2 != nil {
		for _, fn := range a2.Files() {
			defer os.Remove(fn)
		}
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderForName(t *testing.T) {
	if v, ok := providerForName("virtualbox").(*VBoxProvider); !ok {
		t.Fatalf("bad: %#v", v)
	}

	if providerForName("nope") != nil {
		t.Fatal("should be nil if bad provider")
	}
}
