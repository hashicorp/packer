package command

import (
	"path/filepath"
	"testing"
)

func TestCliConfigType(t *testing.T) {
	tc := []struct {
		args     []string
		expected configType
		name     string
	}{
		{
			args: []string{
				filepath.Join(testFixture("build-only"), "template.pkr.json"),
			},
			expected: ConfigTypeJSON,
			name:     "configType: pkr JSON file",
		},
		{
			args: []string{
				filepath.Join(testFixture("build-only"), "template.json"),
			},
			expected: ConfigTypeJSON,
			name:     "configType: JSON file",
		},
		{
			args: []string{
				filepath.Join(testFixture("build-only"), "template.pkr.hcl"),
			},
			expected: ConfigTypeHCL2,
			name:     "configType: HCL2 file",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			c := &BuildCommand{
				Meta: testMetaFile(t),
			}
			cla, _ := c.ParseArgs(tt.args)

			configType, err := cla.MetaArgs.GetConfigType()
			t.Logf("cla: %v", cla)
			if err != nil {
				t.Errorf("Failed to get configType: %v", err)
			}

			if configType != tt.expected {
				t.Errorf("Expected configType: %v; received: %v", ConfigTypeJSON, configType)
			}
		})
	}
}
