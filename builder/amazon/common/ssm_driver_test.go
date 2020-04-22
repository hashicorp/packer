package common

import (
	"testing"
)

func TestStartSession(t *testing.T) {
	tt := []struct {
		Name          string
		PluginName    string
		ErrorExpected bool
	}{
		{"NonExistingPlugin", "boguspluginname", true},
		{"StubExecutablePlugin", "more", false},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			driver := SSMDriver{PluginName: "someboguspluginname"}

			err := driver.StartSession("sessionData", "region", "profile", "params", "bogus-endpoint")

			if tc.ErrorExpected && err == nil {
				t.Fatalf("Executing %q should have failed but instead no error was returned", tc.PluginName)
			}
		})
	}
}
