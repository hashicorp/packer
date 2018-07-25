package vagrant

import (
	"bytes"
	"compress/flate"
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

	config := p.configs[""]
	if config.CompressionLevel != flate.DefaultCompression {
		t.Fatalf("bad: %#v", config.CompressionLevel)
	}

	// Set
	c = testConfig()
	c["compression_level"] = 7
	if err := p.Configure(c); err != nil {
		t.Fatalf("err: %s", err)
	}

	config = p.configs[""]
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

func TestPostProcessorPrepare_subConfigs(t *testing.T) {
	var p PostProcessor

	f, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(f.Name())

	// Default
	c := testConfig()
	c["compression_level"] = 42
	c["vagrantfile_template"] = f.Name()
	c["override"] = map[string]interface{}{
		"aws": map[string]interface{}{
			"compression_level": 7,
		},
	}
	err = p.Configure(c)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.configs[""].CompressionLevel != 42 {
		t.Fatalf("bad: %#v", p.configs[""].CompressionLevel)
	}

	if p.configs[""].VagrantfileTemplate != f.Name() {
		t.Fatalf("bad: %#v", p.configs[""].VagrantfileTemplate)
	}

	if p.configs["aws"].CompressionLevel != 7 {
		t.Fatalf("bad: %#v", p.configs["aws"].CompressionLevel)
	}

	if p.configs["aws"].VagrantfileTemplate != f.Name() {
		t.Fatalf("bad: %#v", p.configs["aws"].VagrantfileTemplate)
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

	if err := os.Remove(name); err != nil {
		t.Fatalf("err: %s", err)
	}

	var p PostProcessor
	if err := p.Configure(c); err == nil {
		t.Fatal("expected an error since vagrantfile_template does not exist")
	}
}

func TestPostProcessorPostProcess_badId(t *testing.T) {
	artifact := &packer.MockArtifact{
		BuilderIdValue: "invalid.packer",
	}

	_, _, err := testPP(t).PostProcess(testUi(), artifact)
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
	a2, _, err := p.PostProcess(testUi(), a)
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
