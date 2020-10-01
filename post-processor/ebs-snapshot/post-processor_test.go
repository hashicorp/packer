package ebssnapshot

import (
	"bytes"
	"context"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testUi() *packer.BasicUi {
	return &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}
}

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packer.PostProcessor = new(PostProcessor)
}

func TestPostProcessor_MapToEC2Tags(t *testing.T) {
	tags := make(map[string]string)
	tags["a"] = "b"
	tags["b"] = "c"

	resultedTags := map_to_ec2_tags(tags)
	if len(resultedTags[0].Tags) != 2 {
		t.Fatal("resulted length is not 2")
	}
}

func TestPostProcessor_PostProcess(t *testing.T) {
	p := &PostProcessor{}
	artifact := &packer.MockArtifact{
		IdValue: "fakeid",
	}

	result, keep, forceOverride, err := p.PostProcess(context.Background(), testUi(), artifact)
	if _, ok := result.(packer.Artifact); !ok {
		t.Fatal("should be instance of Artifact")
	}
	if !keep {
		t.Fatal("should keep")
	}
	if forceOverride {
		t.Fatal("Should default to keep, but not override user wishes")
	}
	if err == nil {
		t.Fatalf("Error should not be nil, invalid volume ID")
	}
}
