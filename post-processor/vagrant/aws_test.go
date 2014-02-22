package vagrant

import (
	"testing"
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
