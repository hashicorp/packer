package client

import "testing"

func TestNormalizeLocation(t *testing.T) {
	tests := []struct {
		name string
		loc  string
		want string
	}{
		{"removes spaces", " with  spaces   ", "withspaces"},
		{"makes lowercase", "MiXed Case", "mixedcase"},
		{"North East US", "North East US", "northeastus"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeLocation(tt.loc); got != tt.want {
				t.Errorf("NormalizeLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}
