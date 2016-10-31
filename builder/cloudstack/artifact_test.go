package cloudstack

import (
	"testing"

	"github.com/mitchellh/packer/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

const templateID = "286dd44a-ec6b-4789-b192-804f08f04b4c"

func TestArtifact_Impl(t *testing.T) {
	var raw interface{} = &Artifact{}

	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact does not implement packer.Artifact")
	}
}

func TestArtifactId(t *testing.T) {
	a := &Artifact{
		client: nil,
		config: nil,
		template: &cloudstack.CreateTemplateResponse{
			Id: "286dd44a-ec6b-4789-b192-804f08f04b4c",
		},
	}

	if a.Id() != templateID {
		t.Fatalf("artifact ID should match: %s", templateID)
	}
}

func TestArtifactString(t *testing.T) {
	a := &Artifact{
		client: nil,
		config: nil,
		template: &cloudstack.CreateTemplateResponse{
			Name: "packer-foobar",
		},
	}
	expected := "A template was created: packer-foobar"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %s", expected)
	}
}
