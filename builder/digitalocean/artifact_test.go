package digitalocean

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact should be artifact")
	}
}

func TestArtifactIdString(t *testing.T) {
	for i, testCase := range []struct {
		artifact       Artifact
		expectedId     string
		expectedString string
	}{
		{
			Artifact{
				droplet: snapshot{"packer-foobar", "42", []string{"sfo"}},
			},
			"sfo:42",
			"A droplet snapshot was created: 'packer-foobar' (ID: 42) in regions 'sfo'",
		},
		{
			Artifact{
				droplet: snapshot{"packer-foobar", "42", []string{"sfo", "tor1"}},
			},
			"sfo,tor1:42",
			"A droplet snapshot was created: 'packer-foobar' (ID: 42) in regions 'sfo,tor1'",
		},
		{
			Artifact{
				droplet: snapshot{"packer-foobar", "42", []string{"sfo", "tor1"}},
				volumes: []snapshot{
					{"packer-foobar-v0", "volid0", []string{"nyc1"}},
				},
			},
			"sfo,tor1:42,nyc1:volid0",
			`A droplet snapshot was created: 'packer-foobar' (ID: 42) in regions 'sfo,tor1'
Volume snapshots were created:
'packer-foobar-v0' (ID: volid0) in regions 'nyc1'`,
		},
		{
			Artifact{
				droplet: snapshot{"packer-foobar", "42", []string{"sfo", "tor1"}},
				volumes: []snapshot{
					{"packer-foobar-v0", "volid0", []string{"nyc1"}},
					{"packer-foobar-v1", "volid1", []string{"nyc1", "nyc3"}},
				},
			},
			"sfo,tor1:42,nyc1:volid0,nyc1,nyc3:volid1",
			`A droplet snapshot was created: 'packer-foobar' (ID: 42) in regions 'sfo,tor1'
Volume snapshots were created:
'packer-foobar-v0' (ID: volid0) in regions 'nyc1'
'packer-foobar-v1' (ID: volid1) in regions 'nyc1,nyc3'`,
		},
	} {
		if e, a := testCase.expectedId, testCase.artifact.Id(); e != a {
			t.Errorf("#%d: expected id %q, but got %q", i, e, a)
		}
		if e, a := testCase.expectedString, testCase.artifact.String(); e != a {
			t.Errorf("#%d: expected string %q, but got %q", i, e, a)
		}
	}
}
