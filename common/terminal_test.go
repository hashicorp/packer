package common

import "testing"

func TestGetTerminalDimensions(t *testing.T) {
	if _, _, err := GetTerminalDimensions(); err != nil {
		t.Fatalf("Unable to get terminal dimensions: %s", err)
	}
}
