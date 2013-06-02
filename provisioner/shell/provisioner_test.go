package shell

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_Defaults(t *testing.T) {
	raw := map[string]interface{}{}

	p := &Provisioner{}
	p.Prepare(raw, nil)

	if p.config.RemotePath != DefaultRemotePath {
		t.Errorf("unexpected remote path: %s", p.config.RemotePath)
	}
}
