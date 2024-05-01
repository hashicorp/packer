package test

import (
	"os"
	"testing"
)

func (ts *PackerTestSuite) TestLoadingOrder() {
	t := ts.T()

	pluginDir := ts.MakePluginDir(t, "1.0.9", "1.0.10")
	defer func() {
		err := os.RemoveAll(pluginDir)
		if err != nil {
			t.Logf("failed to remove temporary plugin directory %q: %s. This may need manual intervention.", pluginDir, err)
		}
	}()

	tests := []struct {
		name         string
		templatePath string
	}{
		{
			"HCL2 - No required_plugins, 1.0.10 is the most recent and should load",
			"./templates/simple.pkr.hcl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.PackerCommand().
				SetArgs("build", tt.templatePath).
				AddEnv("PACKER_PLUGIN_PATH", pluginDir).
				Assert(t, MustSucceed{}, Grep{
					streams: BothStreams,
					expect:  "packer-plugin-tester_v1\\.0\\.10[^\n]+ plugin:",
				})
		})
	}
}
