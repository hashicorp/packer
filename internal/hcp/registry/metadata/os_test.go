package metadata

import (
	"fmt"
	"runtime"
	"testing"
)

// MockExecutor is a mock implementation of CommandExecutor.
type MockExecutor struct {
	stdout string
	err    error
}

// Exec returns a mocked output.
func (m MockExecutor) Exec(name string, arg ...string) ([]byte, error) {
	return []byte(m.stdout), m.err
}

func TestGetInfoForWindows(t *testing.T) {
	tests := []struct {
		name     string
		stdout   string
		err      error
		expected OSInfo
	}{
		{
			name:   "Valid version info",
			stdout: "Microsoft Windows [Version 10.0.19042.928]",
			err:    nil,
			expected: OSInfo{
				Name:    runtime.GOOS,
				Arch:    runtime.GOARCH,
				Version: "10.0.19042.928",
			},
		},
		{
			name:   "Invalid version info",
			stdout: "Invalid output",
			err:    fmt.Errorf("Invalid output"),
			expected: OSInfo{
				Name:    runtime.GOOS,
				Arch:    runtime.GOARCH,
				Version: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockExecutor := MockExecutor{
				stdout: tt.stdout,
				err:    tt.err,
			}

			result := GetInfoForWindows(mockExecutor)

			if result != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}
