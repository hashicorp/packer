package iso

import (
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	commonT "github.com/jetbrains-infra/packer-builder-vsphere/common/testing"
	"testing"
)

func TestBuilderAcc_default(t *testing.T) {
	config := defaultConfig()
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: commonT.RenderConfig(config),
	})
}

func defaultConfig() map[string]interface{} {
	config := map[string]interface{}{
		"vcenter_server":      "vcenter.vsphere65.test",
		"username":            "root",
		"password":            "jetbrains",
		"insecure_connection": true,

		"host": "esxi-1.vsphere65.test",

		"ssh_username": "root",
		"ssh_password": "jetbrains",

		"vm_name":   commonT.NewVMName(),
	}

	return config
}
