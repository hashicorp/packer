package dockertag

import (
	"bytes"
	"context"
	"testing"

	"github.com/hashicorp/packer/builder/docker"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	dockerimport "github.com/hashicorp/packer/post-processor/docker-import"
	"github.com/stretchr/testify/assert"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"repository": "foo",
		"tag":        "bar,buzz",
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

	result, keep, forceOverride, err := p.PostProcess(context.Background(), testUi(), artifact)
	if _, ok := result.(packersdk.Artifact); !ok {
		t.Fatal("should be instance of Artifact")
	}
	if !keep {
		t.Fatal("should keep")
	}
	if !forceOverride {
		t.Fatal("Should force keep no matter what user sets.")
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if driver.TagImageCalled != 2 {
		t.Fatal("should call TagImage")
	}
	if driver.TagImageImageId != "1234567890abcdef" {
		t.Fatal("bad image id")
	}

	if driver.TagImageRepo[0] != "foo:bar" {
		t.Fatal("bad repo")
	}

	if driver.TagImageRepo[1] != "foo:buzz" {
		t.Fatal("bad repo")
	}

	if driver.TagImageForce {
		t.Fatal("bad force. force=false in default")
	}
}

func TestPostProcessor_PostProcess_Force(t *testing.T) {
	driver := &docker.MockDriver{}
	p := &PostProcessor{Driver: driver}
	c := testConfig()
	c["force"] = true
	if err := p.Configure(c); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact := &packer.MockArtifact{
		BuilderIdValue: dockerimport.BuilderId,
		IdValue:        "1234567890abcdef",
	}

	result, keep, forceOverride, err := p.PostProcess(context.Background(), testUi(), artifact)
	if _, ok := result.(packersdk.Artifact); !ok {
		t.Fatal("should be instance of Artifact")
	}
	if !keep {
		t.Fatal("should keep")
	}
	if !forceOverride {
		t.Fatal("Should force keep no matter what user sets.")
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if driver.TagImageCalled != 2 {
		t.Fatal("should call TagImage")
	}
	if driver.TagImageImageId != "1234567890abcdef" {
		t.Fatal("bad image id")
	}
	if driver.TagImageRepo[0] != "foo:bar" {
		t.Fatal("bad repo")
	}
	if driver.TagImageRepo[1] != "foo:buzz" {
		t.Fatal("bad repo")
	}
	if !driver.TagImageForce {
		t.Fatal("bad force")
	}
}

func TestPostProcessor_PostProcess_NoTag(t *testing.T) {
	driver := &docker.MockDriver{}
	p := &PostProcessor{Driver: driver}
	c := testConfig()
	delete(c, "tag")
	if err := p.Configure(c); err != nil {
		t.Fatalf("err %s", err)
	}

	artifact := &packer.MockArtifact{BuilderIdValue: dockerimport.BuilderId, IdValue: "1234567890abcdef"}

	result, keep, forceOverride, err := p.PostProcess(context.Background(), testUi(), artifact)
	if _, ok := result.(packersdk.Artifact); !ok {
		t.Fatal("should be instance of Artifact")
	}
	if !keep {
		t.Fatal("should keep")
	}
	if !forceOverride {
		t.Fatal("Should force keep no matter what user sets.")
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if driver.TagImageCalled != 1 {
		t.Fatal("should call TagImage")
	}
	if driver.TagImageImageId != "1234567890abcdef" {
		t.Fatal("bad image id")
	}
	if driver.TagImageRepo[0] != "foo" {
		t.Fatal("bad repo")
	}
	if driver.TagImageForce {
		t.Fatal("bad force")
	}
}

func TestPostProcessor_PostProcess_Tag_vs_Tags(t *testing.T) {
	testCases := []map[string]interface{}{
		{
			"tag":  "bar,buzz",
			"tags": []string{"bang"},
		},
		{
			"tag":  []string{"bar", "buzz"},
			"tags": []string{"bang"},
		},
		{
			"tag":  []string{"bar"},
			"tags": []string{"buzz", "bang"},
		},
	}

	for _, tc := range testCases {
		var p PostProcessor
		if err := p.Configure(tc); err != nil {
			t.Fatalf("err: %s", err)
		}
		assert.ElementsMatchf(t, p.config.Tags, []string{"bar", "buzz", "bang"},
			"tag and tags fields should be combined into tags fields. Recieved: %#v",
			p.config.Tags)
	}
}
