// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build amd64 && (darwin || windows || linux)

package command

import (
	"log"
	"os"
	"runtime"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/mod/sumdb/dirhash"
)

type testCasePluginsRemove struct {
	name                                    string
	Meta                                    Meta
	inPluginFolder                          map[string]string
	expectedPackerConfigDirHashBeforeRemove string
	packerConfigDir                         string
	pluginSourceArgs                        []string
	want                                    int
	dirFiles                                []string
	expectedPackerConfigDirHashAfterRemove  string
}

func TestPluginsRemoveCommand_Run(t *testing.T) {

	cfg := &configDirSingleton{map[string]string{}}

	tests := []testCasePluginsRemove{
		{
			name: "version-not-installed-no-op",
			Meta: TestMetaFile(t),
			inPluginFolder: map[string]string{
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64":                "1",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM":      "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe":           "1.exe",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM": "07d8453027192ee0c4120242e6e84e2ca2328b8e0f506e2f818a1a5b82790a0b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64":                 "1.out",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM":       "59031c50e0dfeedfde2b4e9445754804dce3f29e4efa737eead0ca9b4f5b85a5",
			},
			expectedPackerConfigDirHashBeforeRemove: "h1:Q5qyAOdD43hL3CquQdVfaHpOYGf0UsZ/+wVA9Ry6cbA=",
			packerConfigDir:                         cfg.dir("1_pkr_plugins_config"),
			pluginSourceArgs:                        []string{"github.com/sylviamoss/comment", "v0.2.19"},
			want:                                    0,
			dirFiles:                                nil,
			expectedPackerConfigDirHashAfterRemove:  "h1:Q5qyAOdD43hL3CquQdVfaHpOYGf0UsZ/+wVA9Ry6cbA=",
		},
		{
			name: "remove-specific-version",
			Meta: TestMetaFile(t),
			inPluginFolder: map[string]string{
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64":                "1",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM":      "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe":           "1.exe",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM": "07d8453027192ee0c4120242e6e84e2ca2328b8e0f506e2f818a1a5b82790a0b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64":                 "1.out",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM":       "59031c50e0dfeedfde2b4e9445754804dce3f29e4efa737eead0ca9b4f5b85a5",
			},
			expectedPackerConfigDirHashBeforeRemove: "h1:Q5qyAOdD43hL3CquQdVfaHpOYGf0UsZ/+wVA9Ry6cbA=",
			packerConfigDir:                         cfg.dir("2_pkr_plugins_config"),
			pluginSourceArgs:                        []string{"github.com/sylviamoss/comment", "v0.2.18"},
			want:                                    0,
			dirFiles: map[string][]string{
				"darwin": {
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM",
				},
				"linux": {
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM",
				},
				"windows": {
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM",
				},
			}[runtime.GOOS],
			expectedPackerConfigDirHashAfterRemove: map[string]string{
				"darwin":  "h1:IMsWPgJZzRhn80t78zE45003gFKN6EXq562/wjaCrKE=",
				"linux":   "h1:Ez7SU1GZLvNGJmoTm9PeFIwHv9fvEgzZAZTMl6874iM=",
				"windows": "h1:RrXlhy9tG9Bi3c2aOzjx/FLLyVNQolcY+MAr4V1etRI=",
			}[runtime.GOOS],
		},
		{
			name: "remove-all-installed-versions",
			Meta: TestMetaFile(t),
			inPluginFolder: map[string]string{
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64":                "1",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM":      "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64":                "1",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64_SHA256SUM":      "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe":           "1.exe",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM": "07d8453027192ee0c4120242e6e84e2ca2328b8e0f506e2f818a1a5b82790a0b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe":           "1.exe",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe_SHA256SUM": "07d8453027192ee0c4120242e6e84e2ca2328b8e0f506e2f818a1a5b82790a0b",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64":                 "1.out",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM":       "59031c50e0dfeedfde2b4e9445754804dce3f29e4efa737eead0ca9b4f5b85a5",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64":                 "1.out",
				"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64_SHA256SUM":       "59031c50e0dfeedfde2b4e9445754804dce3f29e4efa737eead0ca9b4f5b85a5",
			},
			expectedPackerConfigDirHashBeforeRemove: "h1:IEvr6c46+Uk776Hnzy04PuXqnyHGKnnEvIJ713cv0iU=",
			packerConfigDir:                         cfg.dir("2_pkr_plugins_config"),
			pluginSourceArgs:                        []string{"github.com/sylviamoss/comment"},
			want:                                    0,
			dirFiles: map[string][]string{
				"darwin": {
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe_SHA256SUM",
				},
				"linux": {
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe_SHA256SUM",
				},
				"windows": {
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_darwin_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_darwin_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_linux_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_linux_amd64_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.18_x5.0_windows_amd64.exe_SHA256SUM",
					"github.com/sylviamoss/comment/packer-plugin-comment_v0.2.19_x5.0_windows_amd64.exe_SHA256SUM",
				},
			}[runtime.GOOS],
			expectedPackerConfigDirHashAfterRemove: map[string]string{
				"darwin":  "h1:FBBGQ1SKngN9PvF98awv8TZcKaS+CKzJmQoS7vuSXqY=",
				"linux":   "h1:F8lN4Q3sv45ig8r1BLOS/wFuQQy6tSfmuIJf3fnbD5k=",
				"windows": "h1:DOfH6WR1eJNLJcaL8ar8j1xu2WB7Jcn6oG7LGEvNBZI=",
			}[runtime.GOOS],
		},
		{
			name:                                    "no-installed-binaries",
			Meta:                                    TestMetaFile(t),
			inPluginFolder:                          nil,
			expectedPackerConfigDirHashBeforeRemove: "h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
			packerConfigDir:                         cfg.dir("3_pkr_plugins_config"),
			pluginSourceArgs:                        []string{"example.com/sylviamoss/comment", "v0.2.19"},
			want:                                    0,
			dirFiles:                                nil,
			expectedPackerConfigDirHashAfterRemove:  "h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Printf("starting %s", tt.name)
			log.Printf("%#v", tt)
			t.Cleanup(func() {
				_ = os.RemoveAll(tt.packerConfigDir)
			})
			t.Setenv("PACKER_CONFIG_DIR", tt.packerConfigDir)
			createFiles(tt.packerConfigDir, tt.inPluginFolder)

			hash, err := dirhash.HashDir(tt.packerConfigDir, "", dirhash.DefaultHash)
			if err != nil {
				t.Fatalf("HashDir: %v", err)
			}
			if diff := cmp.Diff(tt.expectedPackerConfigDirHashBeforeRemove, hash); diff != "" {
				t.Errorf("unexpected dir hash before plugins remove: +found -expected %s", diff)
			}

			c := &PluginsRemoveCommand{
				Meta: tt.Meta,
			}

			c.CoreConfig.Components.PluginConfig.KnownPluginFolders = []string{tt.packerConfigDir}
			if got := c.Run(tt.pluginSourceArgs); got != tt.want {
				t.Errorf("PluginsRemoveCommand.Run() = %v, want %v", got, tt.want)
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
			if diff := cmp.Diff(tt.expectedPackerConfigDirHashAfterRemove, hash); diff != "" {
				t.Errorf("unexpected dir hash after plugins remove: %s", diff)
			}
		})
	}
}
