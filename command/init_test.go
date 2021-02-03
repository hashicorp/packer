package command

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/mod/sumdb/dirhash"
)

func TestInitCommand_Run(t *testing.T) {
	// These tests will try to optimise for doing the least amount of github api
	// requests whilst testing the max amount of things at once. Hopefully they
	// don't require a GH token just yet. Acc tests are run on linux, darwin and
	// windows, so requests are done 3 times.

	// if os.Getenv(acctest.TestEnvVar) == "" {
	// 	t.Skip(fmt.Sprintf("Acceptance tests skipped unless env '%s' set", acctest.TestEnvVar))
	// }

	type testCase struct {
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
			// here we pre-write plugins with valid checksums, Packer will
			// see those as valid installations it did.
			// the directory hash before and after init should be the same,
			// that's a no-op
			"already-installed-no-op",
			map[string]string{
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64":                 "1",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM":       "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
				"plugin.d/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe":           "1.exe",
				"plugin.d/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM": "07d8453027192ee0c4120242e6e84e2ca2328b8e0f506e2f818a1a5b82790a0b",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64":                  "1.out",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM":        "59031c50e0dfeedfde2b4e9445754804dce3f29e4efa737eead0ca9b4f5b85a5",
			},
			"h1:K4LnrskKho2cVCfhwQ56wS9ZoTVZceWflhc3kmZkCTQ=",
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
			"h1:K4LnrskKho2cVCfhwQ56wS9ZoTVZceWflhc3kmZkCTQ=",
		},
		{
			// here we pre-write plugins with valid checksums, Packer will
			// see those as valid installations it did.
			// But because we require version 0.2.19, we will upgrade.
			"already-installed-upgrade",
			map[string]string{
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64":                 "1",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM":       "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
				"plugin.d/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe":           "1.exe",
				"plugin.d/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM": "07d8453027192ee0c4120242e6e84e2ca2328b8e0f506e2f818a1a5b82790a0b",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64":                  "1.out",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM":        "59031c50e0dfeedfde2b4e9445754804dce3f29e4efa737eead0ca9b4f5b85a5",
			},
			"h1:K4LnrskKho2cVCfhwQ56wS9ZoTVZceWflhc3kmZkCTQ=",
			`# cfg.pkr.hcl
			packer {
				required_plugins {
					comment = {
						source  = "github.com/sylviamoss/comment"
						version = "v0.2.019"
					}
				}
			}`,
			cfg.dir("1"),
			0,
			[]string{
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM",
				"plugin.d/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe",
				"plugin.d/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM",
				map[string]string{
					"darwin":  ".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64_SHA256SUM",
					"linux":   ".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64_SHA256SUM",
					"windows": "plugin.d/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe_SHA256SUM",
				}[runtime.GOOS],
				map[string]string{
					"darwin":  ".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64",
					"linux":   ".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64",
					"windows": "plugin.d/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe",
				}[runtime.GOOS],
			},
			map[string]string{
				"darwin":  "h1:O50ZsZD5Um4/BKSpPWHT0pF6fSYBQgCBxelct9XEBJE=",
				"linux":   "h1:a5F11lgF22JY1MmHRhrzNLnbZYxAZKzCuC9z2trXVMk=",
				"windows": "h1:OmkFZQfC17rHaZFuci+Ib7qChb9Cukzk1VZr3l2dJnw=",
			}[runtime.GOOS],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			pluginDir := ".plugin"
			if runtime.GOOS == "windows" {
				pluginDir = "plugin.d"
			}
			c.CoreConfig.Components.PluginConfig.KnownPluginFolders = []string{filepath.Join(tt.packerConfigDir, pluginDir)}
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
