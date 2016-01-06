package atlas

import (
	"os"
	"reflect"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestPostProcessorConfigure(t *testing.T) {
	currentEnv := os.Getenv("ATLAS_TOKEN")
	os.Setenv("ATLAS_TOKEN", "")
	defer os.Setenv("ATLAS_TOKEN", currentEnv)

	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.client == nil {
		t.Fatal("should have client")
	}
	if p.client.Token != "" {
		t.Fatal("should not have token")
	}
}

func TestPostProcessorConfigure_buildId(t *testing.T) {
	defer os.Setenv(BuildEnvKey, os.Getenv(BuildEnvKey))
	os.Setenv(BuildEnvKey, "5")

	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.buildId != 5 {
		t.Fatalf("bad: %#v", p.config.buildId)
	}
}

func TestPostProcessorConfigure_compileId(t *testing.T) {
	defer os.Setenv(CompileEnvKey, os.Getenv(CompileEnvKey))
	os.Setenv(CompileEnvKey, "5")

	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.compileId != 5 {
		t.Fatalf("bad: %#v", p.config.compileId)
	}
}

func TestPostProcessorMetadata(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	metadata := p.metadata(artifact)
	if len(metadata) > 0 {
		t.Fatalf("bad: %#v", metadata)
	}
}

func TestPostProcessorMetadata_artifact(t *testing.T) {
	config := validDefaults()
	config["metadata"] = map[string]string{
		"foo": "bar",
	}

	var p PostProcessor
	if err := p.Configure(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	artifact.StateValues = map[string]interface{}{
		ArtifactStateMetadata: map[interface{}]interface{}{
			"bar": "baz",
		},
	}

	metadata := p.metadata(artifact)
	expected := map[string]string{
		"foo": "bar",
		"bar": "baz",
	}
	if !reflect.DeepEqual(metadata, expected) {
		t.Fatalf("bad: %#v", metadata)
	}
}

func TestPostProcessorMetadata_config(t *testing.T) {
	config := validDefaults()
	config["metadata"] = map[string]string{
		"foo": "bar",
	}

	var p PostProcessor
	if err := p.Configure(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	metadata := p.metadata(artifact)
	expected := map[string]string{
		"foo": "bar",
	}
	if !reflect.DeepEqual(metadata, expected) {
		t.Fatalf("bad: %#v", metadata)
	}
}

func TestPostProcessorType(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	actual := p.artifactType(artifact)
	if actual != "foo" {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestPostProcessorType_artifact(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(validDefaults()); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := new(packer.MockArtifact)
	artifact.StateValues = map[string]interface{}{
		ArtifactStateType: "bar",
	}
	actual := p.artifactType(artifact)
	if actual != "bar" {
		t.Fatalf("bad: %#v", actual)
	}
}

func validDefaults() map[string]interface{} {
	return map[string]interface{}{
		"artifact":      "mitchellh/test",
		"artifact_type": "foo",
		"test":          true,
	}
}
