package command

import (
	"io/ioutil"
	"path/filepath"
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
		env                                   map[string]string
		want                                  int
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
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64":            "1",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM":  "4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64":           "1.exe",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64_SHA256SUM": "b238233f12d9d803d4df28ac0ce1e80ef93f66ea9391a25ac711a604168472bc",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64":             "1.out",
				".plugin/github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM":   "c28ae3482c9030519a9f5bdf6b3db4638076e6f99897e9b0e71bb38b0d76fd7e",
			},
			"h1:RgZ9LKqioZ4R+GN6oGXpDAMEKreMx1y9uFjyvzVRetI=",
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
			map[string]string{
				"PACKER_CONFIG_DIR": cfg.dir("1"),
			},
			0,
			"h1:RgZ9LKqioZ4R+GN6oGXpDAMEKreMx1y9uFjyvzVRetI=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			c.CoreConfig.Components.PluginConfig.KnownPluginFolders = []string{cfgDir}
			if got := c.Run(args); got != tt.want {
				t.Errorf("InitCommand.Run() = %v, want %v", got, tt.want)
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
