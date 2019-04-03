package dockerpush

import (
	"bytes"
	"context"
	"testing"

	"github.com/hashicorp/packer/builder/docker"
	"github.com/hashicorp/packer/packer"
	dockerimport "github.com/hashicorp/packer/post-processor/docker-import"
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

func TestPostProcessor_PostProcess(t *testing.T) {
	driver := &docker.MockDriver{}
	p := &PostProcessor{Driver: driver}
	artifact := &packer.MockArtifact{
		BuilderIdValue: dockerimport.BuilderId,
		IdValue:        "foo/bar",
	}

	result, keep, err := p.PostProcess(context.Background(), testUi(), artifact)
	if _, ok := result.(packer.Artifact); !ok {
		t.Fatal("should be instance of Artifact")
	}
	if !keep {
		t.Fatal("should keep")
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !driver.PushCalled {
		t.Fatal("should call push")
	}
	if driver.PushName != "foo/bar" {
		t.Fatal("bad name")
	}
	if result.Id() != "foo/bar" {
		t.Fatal("bad image id")
	}
}

func TestPostProcessor_PostProcess_portInName(t *testing.T) {
	driver := &docker.MockDriver{}
	p := &PostProcessor{Driver: driver}
	artifact := &packer.MockArtifact{
		BuilderIdValue: dockerimport.BuilderId,
		IdValue:        "localhost:5000/foo/bar",
	}

	result, keep, err := p.PostProcess(context.Background(), testUi(), artifact)
	if _, ok := result.(packer.Artifact); !ok {
		t.Fatal("should be instance of Artifact")
	}
	if !keep {
		t.Fatal("should keep")
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !driver.PushCalled {
		t.Fatal("should call push")
	}
	if driver.PushName != "localhost:5000/foo/bar" {
		t.Fatal("bad name")
	}
	if result.Id() != "localhost:5000/foo/bar" {
		t.Fatal("bad image id")
	}
}

func TestPostProcessor_PostProcess_tags(t *testing.T) {
	driver := &docker.MockDriver{}
	p := &PostProcessor{Driver: driver}
	artifact := &packer.MockArtifact{
		BuilderIdValue: dockerimport.BuilderId,
		IdValue:        "hashicorp/ubuntu:precise",
	}

	result, keep, err := p.PostProcess(context.Background(), testUi(), artifact)
	if _, ok := result.(packer.Artifact); !ok {
		t.Fatal("should be instance of Artifact")
	}
	if !keep {
		t.Fatal("should keep")
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !driver.PushCalled {
		t.Fatal("should call push")
	}
	if driver.PushName != "hashicorp/ubuntu:precise" {
		t.Fatalf("bad name: %s", driver.PushName)
	}
	if result.Id() != "hashicorp/ubuntu:precise" {
		t.Fatal("bad image id")
	}
}
