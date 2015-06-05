package dockerfile

import (
	"bytes"
	"testing"

	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/post-processor/docker-import"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"maintainer": "foo",
		"cmd":        []interface{}{ "/foo/bar" },
		"label":      map[string]string{ "foo": "bar" },
		"expose":     []string{ "1234" },
		"env":        map[string]string{ "foo": "bar" },
		"entrypoint": []interface{}{ "/foo/bar" },
		"volume":     []string{ "/foo/bar" },
		"user":       "foo",
		"workdir":    "/foo/bar",
	}
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

func TestPostProcessor_PostProcess(t *testing.T) {
	driver := &docker.MockDriver{}
	p := &PostProcessor{Driver: driver}
	if err := p.Configure(testConfig()); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := &packer.MockArtifact{
		BuilderIdValue: dockerimport.BuilderId,
		IdValue:        "1234567890abcdef",
	}

	result, keep, err := p.PostProcess(testUi(), artifact)
	if _, ok := result.(packer.Artifact); !ok {
		t.Fatal("should be instance of Artifact")
	}
	if !keep {
		t.Fatal("should keep")
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !driver.BuildImageCalled {
		t.Fatal("should call BuildImage")
	}

	dockerfile := `FROM 1234567890abcdef
MAINTAINER foo
CMD ["/foo/bar"]
LABEL "foo"="bar"
EXPOSE 1234
ENV foo bar
ENTRYPOINT ["/foo/bar"]
VOLUME ["/foo/bar"]
USER foo
WORKDIR /foo/bar`

	if driver.BuildImageDockerfile.String() != dockerfile {
		t.Fatalf("should render Dockerfile correctly: %s", driver.BuildImageDockerfile.String())
	}
}

func TestPostProcessor_processVar(t *testing.T) {
	driver := &docker.MockDriver{}
	p := &PostProcessor{Driver: driver}
	if err := p.Configure(testConfig()); err != nil {
		t.Fatalf("err: %s", err)
	}

	res, err := p.processVar("foo");
	if err != nil {
		t.Fatalf("failed to process variable: %s", err)
	}
	if res != "foo" {
		t.Fatalf("should be foo: %s", res)
	}

	res, err = p.processVar([]string{ "foo", "bar" });
	if err != nil {
		t.Fatalf("failed to process variable: %s", err)
	}
	if res != `["foo","bar"]` {
		t.Fatalf(`should be ["foo","bar"]: %s`, res)
	}

	res, err = p.processVar([]interface{}{ "foo", "bar" });
	if err != nil {
		t.Fatalf("failed to process variable: %s", err)
	}
	if res != `["foo","bar"]` {
		t.Fatalf(`should be ["foo","bar"]: %s`, res)
	}

	_, err = p.processVar(nil);
	if err != nil {
		t.Fatalf("failed to process variable: %s", err)
	}
}
