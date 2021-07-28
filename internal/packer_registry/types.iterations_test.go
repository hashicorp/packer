package packer_registry

import (
	"os"
	"path"
	"testing"

	git "github.com/go-git/go-git/v5"
)

func TestNewIteration(t *testing.T) {
	var tc = []struct {
		name          string
		fingerprint   string
		opts          IterationOptions
		setupFn       func() func()
		errorExpected bool
	}{
		{
			name:        "Using Fingerprint Env variable",
			fingerprint: "6825d1ad0d5e",
			setupFn: func() func() {
				os.Setenv("HCP_PACKER_BUILD_FINGERPRINT", "6825d1ad0d5e")
				return func() {
					os.Unsetenv("HCP_PACKER_BUILD_FINGERPRINT")
				}
			},
		},
		{
			name:        "Using Git Fingerprint",
			fingerprint: "4ec004e18e977a5b8a3a28f4b24279b6993d7e7c",
			setupFn: func() func() {
				//no:lint
				git.PlainClone(tempdir("4ec004e18e"), false, &git.CloneOptions{
					// Archived repo
					URL:   "https://github.com/hashicorp/packer-builder-vsphere",
					Depth: 1,
				})

				return func() {
					//no:lint
					os.RemoveAll(tempdir("4ec004e18e"))
				}

			},
			opts: IterationOptions{
				TemplateBaseDir: tempdir("4ec004e18e"),
			},
		},
		{
			name: "Using No Fingerprint",
			opts: IterationOptions{
				TemplateBaseDir: "/dev/null",
			},
			errorExpected: true,
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFn != nil {
				cleanup := tt.setupFn()
				defer cleanup()
			}

			i, err := NewIteration(tt.opts)
			if tt.errorExpected && err != nil {
				t.Logf("the expected error is %q", err)
				return
			}

			if tt.errorExpected && err == nil {
				t.Errorf("expected %q to fail, but it didn't", tt.name)
			}

			if i.Fingerprint != tt.fingerprint {
				t.Errorf("%q failed to load the expected fingerprint %q, but got %q", tt.name, tt.fingerprint, i.Fingerprint)
			}

		})
	}
}

func tempdir(dirname string) string {
	return path.Join(os.TempDir(), dirname)
}
