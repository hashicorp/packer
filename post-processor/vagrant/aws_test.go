package vagrant

import (
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestAWSProvider_impl(t *testing.T) {
	var _ Provider = new(AWSProvider)
}

func TestAWSProvider_KeepInputArtifact(t *testing.T) {
	p := new(AWSProvider)

	if !p.KeepInputArtifact() {
		t.Fatal("should keep input artifact")
	}
}

func TestAWSProvider_ArtifactId(t *testing.T) {
	p := new(AWSProvider)
	ui := testUi()
	artifact := &packersdk.MockArtifact{
		IdValue: "us-east-1:ami-1234",
	}

	vagrantfile, _, err := p.Process(ui, artifact, "foo")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	result := `aws.region_config "us-east-1", ami: "ami-1234"`
	if !strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
}
