package packer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCachePath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	tmp := os.TempDir()

	// reset env
	cd := os.Getenv("PACKER_CACHE_DIR")
	os.Setenv("PACKER_CACHE_DIR", "")
	defer func() {
		os.Setenv("PACKER_CACHE_DIR", cd)
	}()

	type args struct {
		paths []string
	}
	tests := []struct {
		name    string
		args    args
		env     map[string]string
		want    string
		wantErr bool
	}{
		{"base", args{}, nil, filepath.Join(wd, "packer_cache"), false},
		{"base and path", args{[]string{"a", "b"}}, nil, filepath.Join(wd, "packer_cache", "a", "b"), false},
		{"env and path", args{[]string{"a", "b"}}, map[string]string{"PACKER_CACHE_DIR": tmp}, filepath.Join(tmp, "a", "b"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				os.Setenv(k, v)
			}
			got, err := CachePath(tt.args.paths...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CachePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CachePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
