package registry

import (
	"os"
	"path"
	"testing"

	git "github.com/go-git/go-git/v5"
)

func TestIteration_Initialize(t *testing.T) {
	var tc = []struct {
		name          string
		fingerprint   string
		setupFn       func(t *testing.T)
		errorExpected bool
	}{
		{
			name:        "using fingerprint env variable",
			fingerprint: "6825d1ad0d5e",
			setupFn: func(t *testing.T) {
				t.Setenv("HCP_PACKER_BUILD_FINGERPRINT", "6825d1ad0d5e")
			},
		},
		{
			name:        "using git fingerprint",
			fingerprint: "4ec004e18e977a5b8a3a28f4b24279b6993d7e7c",
			setupFn: func(t *testing.T) {
				//nolint:errcheck
				git.PlainClone(tempdir("4ec004e18e"), false, &git.CloneOptions{
					// Archived repo
					URL:   "https://github.com/hashicorp/packer-builder-vsphere",
					Depth: 1,
				})

				t.Setenv("HCP_PACKER_BUILD_FINGERPRINT", "4ec004e18e977a5b8a3a28f4b24279b6993d7e7c")

				t.Cleanup(func() {
					//nolint:errcheck
					os.RemoveAll(tempdir("4ec004e18e"))
				})
			},
		},
		{
			name: "using no fingerprint in clean directory",
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			i := NewIteration()
			err := i.Initialize()
			if tt.errorExpected {
				t.Logf("%v", err)
				if err == nil {
					t.Errorf("expected %q to result in an error, but it return no error", tt.name)
				}

				if i.Fingerprint != "" {
					t.Errorf("expected %q to result in an error with an empty iteration fingerprint, but got %q", tt.name, i.Fingerprint)
				}
				return
			}

			if err != nil {
				t.Errorf("expected %q to return with no error, but it %v", tt.name, err)
			}

			if tt.fingerprint != "" && i.Fingerprint != tt.fingerprint {
				t.Errorf("%q failed to load the expected fingerprint %q, but got %q", tt.name, tt.fingerprint, i.Fingerprint)
			}

		})
	}
}

func tempdir(dirname string) string {
	return path.Join(os.TempDir(), dirname)
}
