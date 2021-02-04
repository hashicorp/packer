package command

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"golang.org/x/mod/sumdb/dirhash"
)

func TestInitCommand_Run(t *testing.T) {
	// These tests will try to optimise for doing the least amount of github api
	// requests whilst testing the max amount of things at once. Hopefully they
	// don't require a GH token just yet. Acc tests are run on linux, darwin and
	// windows, so requests are done 3 times.

	type testCase struct {
		checkSkip                             func(*testing.T)
		name                                  string
		inPluginFolder                        map[string]string
		expectedPackerConfigDirHashBeforeInit string
		hclFile                               string
		packerConfigDir                       string
		want                                  int
		dirFiles                              []string
		expectedPackerConfigDirHashAfterInit  string
	}

	cfg := &configDirSingleton{map[string]string{}}

	tests := []testCase{
		{
			nil,
			// here we pre-write plugins with valid checksums, Packer will
			// see those as valid installations it did.
			// the directory hash before and after init should be the same,
			// that's a no-op. This also should do no GH query, so it is best
			// to always run it.
			"already-installed-no-op",
			map[string]string{
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64":                "1",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM":      "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe":           "1.exe",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM": "07d8453027192ee0c4120242e6e84e2ca2328b8e0f506e2f818a1a5b82790a0b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64":                 "1.out",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM":       "59031c50e0dfeedfde2b4e9445754804dce3f29e4efa737eead0ca9b4f5b85a5",
			},
			"h1:Q5qyAOdD43hL3CquQdVfaHpOYGf0UsZ/+wVA9Ry6cbA=",
			`# cfg.pkr.hcl
			packer {
				required_plugins {
					comment = {
						source  = "github.com/sylviamoss/comment"
						version = "v0.2.018"
					}
				}
			}`,
			cfg.dir("1"),
			0,
			nil,
			"h1:Q5qyAOdD43hL3CquQdVfaHpOYGf0UsZ/+wVA9Ry6cbA=",
		},
		{
			func(t *testing.T) {
				if os.Getenv(acctest.TestEnvVar) == "" {
					t.Skipf("Acceptance test skipped unless env '%s' set", acctest.TestEnvVar)
				}
			},
			// here we pre-write plugins with valid checksums, Packer will
			// see those as valid installations it did.
			// But because we require version 0.2.19, we will upgrade.
			"already-installed-upgrade",
			map[string]string{
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64":                "1",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM":      "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe":           "1.exe",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM": "07d8453027192ee0c4120242e6e84e2ca2328b8e0f506e2f818a1a5b82790a0b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64":                 "1.out",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM":       "59031c50e0dfeedfde2b4e9445754804dce3f29e4efa737eead0ca9b4f5b85a5",
			},
			"h1:Q5qyAOdD43hL3CquQdVfaHpOYGf0UsZ/+wVA9Ry6cbA=",
			`# cfg.pkr.hcl
			packer {
				required_plugins {
					comment = {
						source  = "github.com/sylviamoss/comment"
						version = "v0.2.019"
					}
				}
			}`,
			cfg.dir("2"),
			0,
			[]string{
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM",
				map[string]string{
					"darwin":  "github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64_SHA256SUM",
					"linux":   "github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64_SHA256SUM",
					"windows": "github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe_SHA256SUM",
				}[runtime.GOOS],
				map[string]string{
					"darwin":  "github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64",
					"linux":   "github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64",
					"windows": "github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe",
				}[runtime.GOOS],
			},
			map[string]string{
				"darwin":  "h1:ORwcCYUx8z/5n/QvuTJo2vrgKpfJA4AxlNg1G9/BCDI=",
				"linux":   "h1:CGym0+Nd0LEANgzqL0wx/LDjRL8bYwlpZ0HajPJo/hs=",
				"windows": "h1:ag0/C1YjP7KoEDYOiJHE0K+lhFgs0tVgjriWCXVT1fg=",
			}[runtime.GOOS],
		},
		{
			func(t *testing.T) {
				if os.Getenv(acctest.TestEnvVar) == "" {
					t.Skipf("Acceptance test skipped unless env '%s' set", acctest.TestEnvVar)
				}
			},
			"release-with-no-binary",
			nil,
			"h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
			`# cfg.pkr.hcl
			packer {
				required_plugins {
					comment = {
						source  = "github.com/sylviamoss/comment"
						version = "v0.2.20"
					}
				}
			}`,
			cfg.dir("3"),
			1,
			nil,
			"h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.checkSkip != nil {
				tt.checkSkip(t)
				if t.Skipped() {
					return
				}
			}
			log.Printf("starting %s", tt.name)
			createFiles(tt.packerConfigDir, tt.inPluginFolder)

			hash, err := dirhash.HashDir(tt.packerConfigDir, "", dirhash.DefaultHash)
			if err != nil {
				t.Fatalf("HashDir: %v", err)
			}
			if diff := cmp.Diff(tt.expectedPackerConfigDirHashBeforeInit, hash); diff != "" {
				t.Errorf("unexpected dir hash before init: %s", diff)
			}

			cfgDir, err := ioutil.TempDir("", "pkr-test-init-file-folder")
			if err != nil {
				t.Fatalf("TempDir: %v", err)
			}
			if err := ioutil.WriteFile(filepath.Join(cfgDir, "cfg.pkr.hcl"), []byte(tt.hclFile), 0666); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			args := []string{cfgDir}

			c := &InitCommand{
				Meta: testMetaFile(t),
			}

			c.CoreConfig.Components.PluginConfig.KnownPluginFolders = []string{tt.packerConfigDir}
			if got := c.Run(args); got != tt.want {
				t.Errorf("InitCommand.Run() = %v, want %v", got, tt.want)
			}

			if tt.dirFiles != nil {
				dirFiles, err := dirhash.DirFiles(tt.packerConfigDir, "")
				if err != nil {
					t.Fatalf("DirFiles: %v", err)
				}
				sort.Strings(tt.dirFiles)
				sort.Strings(dirFiles)
				if diff := cmp.Diff(tt.dirFiles, dirFiles); diff != "" {
					t.Errorf("found files differ: %v", diff)
				}
			}

			hash, err = dirhash.HashDir(tt.packerConfigDir, "", dirhash.DefaultHash)
			if err != nil {
				t.Fatalf("HashDir: %v", err)
			}
			if diff := cmp.Diff(tt.expectedPackerConfigDirHashAfterInit, hash); diff != "" {
				t.Errorf("unexpected dir hash after init: %s", diff)
			}
		})
	}
}
