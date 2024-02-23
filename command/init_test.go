// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build amd64 && (darwin || windows || linux)

package command

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-getter/v2"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"golang.org/x/mod/sumdb/dirhash"
)

type testCaseInit struct {
	name                                  string
	setup                                 []func(*testing.T, testCaseInit)
	Meta                                  Meta
	inPluginFolder                        map[string]string
	expectedPackerConfigDirHashBeforeInit string
	inConfigFolder                        map[string]string
	packerConfigDir                       string
	packerUserFolder                      string
	want                                  int
	dirFiles                              []string
	expectedPackerConfigDirHashAfterInit  string
	moreTests                             []func(*testing.T, testCaseInit)
}

type testBuild struct {
	want int
}

func (tb testBuild) fn(t *testing.T, tc testCaseInit) {
	bc := BuildCommand{
		Meta: tc.Meta,
	}

	args := []string{tc.packerUserFolder}
	want := tb.want
	if got := bc.Run(args); got != want {
		t.Errorf("BuildCommand.Run() = %v, want %v", got, want)
	}
}

func TestInitCommand_Run(t *testing.T) {
	// These tests will try to optimise for doing the least amount of github api
	// requests whilst testing the max amount of things at once. Hopefully they
	// don't require a GH token just yet. Acc tests are run on linux, darwin and
	// windows, so requests are done 3 times.

	cfg := &configDirSingleton{map[string]string{}}

	tests := []testCaseInit{
		//	{
		//		// here we pre-write plugins with valid checksums, Packer will
		//		// see those as valid installations it did.
		//		// the directory hash before and after init should be the same,
		//		// that's a no-op. This also should do no GH query, so it is best
		//		// to always run it.
		//		//
		//		// Note: cannot work with plugin changes since the fake binary
		//		// isn't recognised  as a potential plugin, so Packer always
		//		// installs it.
		//		"already-installed-no-op",
		//		nil,
		//		TestMetaFile(t),
		//		map[string]string{
		//			"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_darwin_amd64":                "1",
		//			"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_darwin_amd64_SHA256SUM":      "a23e48324f2d9b912a89354945562b21b0ae99133b31d3132e2e6671aba8e085",
		//			"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_windows_amd64.exe":           "1.exe",
		//			"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_windows_amd64.exe_SHA256SUM": "f1cf5865b35933b8e5195625ac8be44487b64007f223912cc5c1784e493e62b2",
		//			"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_linux_amd64":                 "1.out",
		//			"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_linux_amd64_SHA256SUM":       "0a4e4e1d6de28054f64946782a5eb92edc663e980ae0780fcb3a614d27c58506",
		//		},
		//		"h1:jQchMpyaQhkZYn0iguw6E6O4VCWxacYx2aR/RJJNLmo=",
		//		map[string]string{
		//			`cfg.pkr.hcl`: `
		//				packer {
		//					required_plugins {
		//						comment = {
		//							source  = "github.com/hashicorp/hashicups"
		//							version = "v1.0.1"
		//						}
		//					}
		//				}`,
		//		},
		//		cfg.dir("1_pkr_config"),
		//		cfg.dir("1_pkr_user_folder"),
		//		0,
		//		nil,
		//		"h1:jQchMpyaQhkZYn0iguw6E6O4VCWxacYx2aR/RJJNLmo=",
		//		[]func(t *testing.T, tc testCaseInit){
		//			// test that a build will not work since plugins are broken for
		//			// this tests (they are not binaries).
		//			testBuild{want: 1}.fn,
		//		},
		//	},
		{
			// here we pre-write plugins with valid checksums, Packer will
			// see those as valid installations it did.
			// But because we require version 1.0.2, we will upgrade.
			"already-installed-upgrade",
			[]func(t *testing.T, tc testCaseInit){
				skipInitTestUnlessEnVar(acctest.TestEnvVar).fn,
			},
			TestMetaFile(t),
			map[string]string{
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_darwin_amd64":                "1",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_darwin_amd64_SHA256SUM":      "a23e48324f2d9b912a89354945562b21b0ae99133b31d3132e2e6671aba8e085",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_windows_amd64.exe":           "1.exe",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_windows_amd64.exe_SHA256SUM": "f1cf5865b35933b8e5195625ac8be44487b64007f223912cc5c1784e493e62b2",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_linux_amd64":                 "1.out",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_linux_amd64_SHA256SUM":       "0a4e4e1d6de28054f64946782a5eb92edc663e980ae0780fcb3a614d27c58506",
			},
			"h1:jQchMpyaQhkZYn0iguw6E6O4VCWxacYx2aR/RJJNLmo=",
			map[string]string{
				`cfg.pkr.hcl`: `
					packer {
						required_plugins {
							hashicups = {
								source  = "github.com/hashicorp/hashicups"
								version = "v1.0.2"
							}
						}
					}`,
				`source.pkr.hcl`: `
				source "null" "test" {
					communicator = "none"
				}
				`,
				`build.pkr.hcl`: `
				build {
					sources = ["null.test"]
					provisioner "hashicups-toppings" {
						toppings = ["sugar"] # Takes 5 seconds in the current state
					}
				}
				`,
			},
			cfg.dir("2_pkr_config"),
			cfg.dir("2_pkr_user_folder"),
			0,
			[]string{
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_darwin_amd64",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_darwin_amd64_SHA256SUM",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_linux_amd64",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_linux_amd64_SHA256SUM",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_windows_amd64.exe",
				"github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.1_x5.0_windows_amd64.exe_SHA256SUM",
				map[string]string{
					"darwin":  "github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.2_x5.0_darwin_amd64_SHA256SUM",
					"linux":   "github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.2_x5.0_linux_amd64_SHA256SUM",
					"windows": "github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.2_x5.0_windows_amd64.exe_SHA256SUM",
				}[runtime.GOOS],
				map[string]string{
					"darwin":  "github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.2_x5.0_darwin_amd64",
					"linux":   "github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.2_x5.0_linux_amd64",
					"windows": "github.com/hashicorp/hashicups/packer-plugin-hashicups_v1.0.2_x5.0_windows_amd64.exe",
				}[runtime.GOOS],
			},
			map[string]string{
				"darwin":  "h1:ptsMLvUeLsMMeXDJP2PWKAKIkE+kWVhOkhNYOYPJbSE=",
				"linux":   "h1:ivCmyQ+/qNXfBsyeccGsa7P5232q7MUZk83B3yl80Ms=",
				"windows": "h1:BeqAUnyGiBg9fVuf9Cn9a4h91bgdZ2U4kV7EuQKefcM=",
			}[runtime.GOOS],
			[]func(t *testing.T, tc testCaseInit){
				// test that a build will work as the plugin was just installed
				testBuild{want: 0}.fn,
			},
		},
		{
			"release-with-no-binary",
			[]func(t *testing.T, tc testCaseInit){
				skipInitTestUnlessEnVar(acctest.TestEnvVar).fn,
			},
			TestMetaFile(t),
			nil,
			"h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
			map[string]string{
				`cfg.pkr.hcl`: `
					packer {
						required_plugins {
							comment = {
								source  = "github.com/sylviamoss/comment"
								version = "v0.2.20"
							}
						}
					}`,
			},
			cfg.dir("3_pkr_config"),
			cfg.dir("3_pkr_user_folder"),
			1,
			nil,
			"h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
			nil,
		},
		{
			"unsupported-non-github-source-address",
			[]func(t *testing.T, tc testCaseInit){
				skipInitTestUnlessEnVar(acctest.TestEnvVar).fn,
			},
			TestMetaFile(t),
			nil,
			"h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
			map[string]string{
				`cfg.pkr.hcl`: `
					packer {
						required_plugins {
							comment = {
								source  = "example.com/sylviamoss/comment"
								version = "v0.2.19"
							}
						}
					}`,
			},
			cfg.dir("6_pkr_config"),
			cfg.dir("6_pkr_user_folder"),
			1,
			nil,
			"h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Printf("starting %s", tt.name)
			log.Printf("%#v", tt)
			t.Cleanup(func() {
				_ = os.RemoveAll(tt.packerConfigDir)
			})
			t.Cleanup(func() {
				_ = os.RemoveAll(tt.packerUserFolder)
			})
			t.Setenv("PACKER_CONFIG_DIR", tt.packerConfigDir)
			for _, init := range tt.setup {
				init(t, tt)
				if t.Skipped() {
					return
				}
			}
			createFiles(tt.packerConfigDir, tt.inPluginFolder)
			createFiles(tt.packerUserFolder, tt.inConfigFolder)

			hash, err := dirhash.HashDir(tt.packerConfigDir, "", dirhash.DefaultHash)
			if err != nil {
				t.Fatalf("HashDir: %v", err)
			}
			if diff := cmp.Diff(tt.expectedPackerConfigDirHashBeforeInit, hash); diff != "" {
				t.Errorf("unexpected dir hash before init: +found -expected %s", diff)
			}

			args := []string{tt.packerUserFolder}

			c := &InitCommand{
				Meta: tt.Meta,
			}

			if err := c.CoreConfig.Components.PluginConfig.Discover(); err != nil {
				t.Fatalf("Failed to discover plugins: %s", err)
			}

			c.CoreConfig.Components.PluginConfig.PluginDirectory = tt.packerConfigDir
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

			for i, subTest := range tt.moreTests {
				t.Run(fmt.Sprintf("-subtest-%d", i), func(t *testing.T) {
					subTest(t, tt)
				})
			}
		})
	}
}

type skipInitTestUnlessEnVar string

func (key skipInitTestUnlessEnVar) fn(t *testing.T, tc testCaseInit) {
	// always run acc tests for now
	// if os.Getenv(string(key)) == "" {
	// 	t.Skipf("Acceptance test skipped unless env '%s' set", key)
	// }
}

type initTestGoGetPlugin struct {
	Src string
	Dst string
}

func (opts initTestGoGetPlugin) fn(t *testing.T, _ testCaseInit) {
	if _, err := getter.Get(context.Background(), opts.Dst, opts.Src); err != nil {
		t.Fatalf("get: %v", err)
	}
}

// TestInitCmd aims to test the init command, with output validation
func TestInitCmd(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCode int
		outputCheck  func(string, string) error
	}{
		{
			name: "Ensure init warns on template without required_plugin blocks",
			args: []string{
				testFixture("hcl", "build-var-in-pp.pkr.hcl"),
			},
			expectedCode: 0,
			outputCheck: func(stdout, stderr string) error {
				if !strings.Contains(stdout, "No plugins requirement found") {
					return fmt.Errorf("command should warn about plugin requirements not found, but did not")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &InitCommand{
				Meta: TestMetaFile(t),
			}

			exitCode := c.Run(tt.args)
			if exitCode != tt.expectedCode {
				t.Errorf("process exit code mismatch: expected %d, got %d",
					tt.expectedCode,
					exitCode)
			}

			out, stderr := GetStdoutAndErrFromTestMeta(t, c.Meta)
			err := tt.outputCheck(out, stderr)
			if err != nil {
				if len(out) != 0 {
					t.Logf("command stdout: %q", out)
				}

				if len(stderr) != 0 {
					t.Logf("command stderr: %q", stderr)
				}
				t.Error(err.Error())
			}
		})
	}
}
