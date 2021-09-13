package registry

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
			name:        "using fingerprint env variable",
			fingerprint: "6825d1ad0d5e",
			setupFn: func() func() {
				os.Setenv("HCP_PACKER_BUILD_FINGERPRINT", "6825d1ad0d5e")
				return func() {
					os.Unsetenv("HCP_PACKER_BUILD_FINGERPRINT")
				}
			},
		},
		{
			name:        "using git fingerprint",
			fingerprint: "4ec004e18e977a5b8a3a28f4b24279b6993d7e7c",
			opts: IterationOptions{
				TemplateBaseDir: tempdir("4ec004e18e"),
			},
			setupFn: func() func() {
				//nolint:errcheck
				git.PlainClone(tempdir("4ec004e18e"), false, &git.CloneOptions{
					// Archived repo
					URL:   "https://github.com/hashicorp/packer-builder-vsphere",
					Depth: 1,
				})

				return func() {
					//nolint:errcheck
					os.RemoveAll(tempdir("4ec004e18e"))
				}
			},
		},
		{
			name: "using empty git directory",
			opts: IterationOptions{
				TemplateBaseDir: tempdir("empty-init"),
			},
			setupFn: func() func() {
				//nolint:errcheck
				git.PlainInit(tempdir("empty-init"), false)
				return func() {
					//nolint:errcheck
					os.RemoveAll(tempdir("empty-init"))
				}
			},
			errorExpected: true,
		},
		{
			name: "using no fingerprint in clean directory",
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
			if tt.errorExpected {
				t.Logf("%v", err)
				if err == nil {
					t.Errorf("expected %q to result in an error, but it return no error", tt.name)
				}

				if i != nil {
					t.Errorf("expected %q to result in an error with no iteration, but got %v", tt.name, i)
				}
				return
			}

			if err != nil {
				t.Errorf("expected %q to return with no error, but it %v", tt.name, err)
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
