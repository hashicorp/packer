package test

import (
	"testing"
)

func (ts *PackerTestSuite) TestLoadingOrder() {
	t := ts.T()

	pluginDir, cleanup := ts.MakePluginDir("1.0.9", "1.0.10")
	defer cleanup()

	for _, command := range []string{"build", "validate"} {
		tests := []struct {
			name         string
			templatePath string
			grepStr      string
		}{
			{
				"HCL2 " + command + " - With required_plugins, 1.0.10 is the most recent and should load",
				"./templates/simple.pkr.hcl",
				"packer-plugin-tester_v1\\.0\\.10[^\n]+ plugin:",
			},
			{
				"JSON " + command + " - No required_plugins, 1.0.10 is the most recent and should load",
				"./templates/simple.json",
				"packer-plugin-tester_v1\\.0\\.10[^\n]+ plugin:",
			},
			{
				"HCL2 " + command + " - With required_plugins, 1.0.9 is pinned, so 1.0.9 should be used",
				"./templates/pin_1.0.9.pkr.hcl",
				"packer-plugin-tester_v1\\.0\\.9[^\n]+ plugin:",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ts.PackerCommand().
					SetArgs(command, tt.templatePath).
					UsePluginDir(pluginDir).
					Assert(t, MustSucceed{}, Grep{
						streams: BothStreams,
						expect:  tt.grepStr,
					})
			})
		}
	}
}
