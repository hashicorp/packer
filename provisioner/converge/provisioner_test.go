package converge

import (
	"strings"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"bootstrap": false,
		"version":   "",
		"module_dirs": []map[string]interface{}{
			{
				"source":      "from",
				"destination": "/opt/converge",
			},
		},
		"modules": []map[string]interface{}{
			{
				"module": "/opt/converge/test.hcl",
			},
		},
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatal("must be a Provisioner")
	}
}

func TestProvisionerPrepare(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		var p Provisioner
		config := testConfig()

		// delete any keys that we're testing here to make sure they're actually
		// being set by `Prepare`
		delete(config["modules"].([]map[string]interface{})[0], "directory")

		if err := p.Prepare(config); err != nil {
			t.Errorf("err: %s", err)
		}

		if p.config.Modules[0].WorkingDirectory != "/tmp" {
			t.Errorf("unexpected module directory: %s", p.config.Modules[0].WorkingDirectory)
		}
	})

	t.Run("validate", func(t *testing.T) {
		t.Run("bad version", func(t *testing.T) {
			var p Provisioner
			config := testConfig()
			config["version"] = "bad version with spaces"

			err := p.Prepare(config)
			if err == nil {
				t.Error("expected error")
			} else if !strings.HasPrefix(err.Error(), "Invalid Converge version") {
				t.Errorf("expected error starting with \"Invalid Converge version\". Got: %s", err)
			}
		})

		t.Run("module dir", func(t *testing.T) {
			t.Run("missing source", func(t *testing.T) {
				var p Provisioner
				config := testConfig()
				delete(config["module_dirs"].([]map[string]interface{})[0], "source")

				err := p.Prepare(config)
				if err == nil {
					t.Error("expected error")
				} else if err.Error() != "Source (\"source\" key) is required in Converge module dir #0" {
					t.Errorf("bad error message: %s", err)
				}
			})

			t.Run("missing destination", func(t *testing.T) {
				var p Provisioner
				config := testConfig()
				delete(config["module_dirs"].([]map[string]interface{})[0], "destination")

				err := p.Prepare(config)
				if err == nil {
					t.Error("expected error")
				} else if err.Error() != "Destination (\"destination\" key) is required in Converge module dir #0" {
					t.Errorf("bad error message: %s", err)
				}
			})
		})

		t.Run("modules", func(t *testing.T) {
			t.Run("none specified", func(t *testing.T) {
				var p Provisioner
				config := testConfig()
				delete(config, "modules")

				err := p.Prepare(config)
				if err == nil {
					t.Error("expected error")
				} else if err.Error() != "Converge requires at least one module (\"modules\" key) to provision the system" {
					t.Errorf("bad error message: %s", err)
				}
			})

			t.Run("missing module", func(t *testing.T) {
				var p Provisioner
				config := testConfig()
				delete(config["modules"].([]map[string]interface{})[0], "module")

				err := p.Prepare(config)
				if err == nil {
					t.Error("expected error")
				} else if err.Error() != "Module (\"module\" key) is required in Converge module #0" {
					t.Errorf("bad error message: %s", err)
				}
			})
		})
	})
}
