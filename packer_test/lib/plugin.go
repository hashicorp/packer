package lib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

// LDFlags compiles the ldflags for the plugin to compile based on the information provided.
func LDFlags(version *version.Version) string {
	pluginPackage := "github.com/hashicorp/packer-plugin-tester"

	ldflagsArg := fmt.Sprintf("-X %s/version.Version=%s", pluginPackage, version.Core())
	if version.Prerelease() != "" {
		ldflagsArg = fmt.Sprintf("%s -X %s/version.VersionPrerelease=%s", ldflagsArg, pluginPackage, version.Prerelease())
	}
	if version.Metadata() != "" {
		ldflagsArg = fmt.Sprintf("%s -X %s/version.VersionMetadata=%s", ldflagsArg, pluginPackage, version.Metadata())
	}

	return ldflagsArg
}

// BinaryName is the raw name of the plugin binary to produce
//
// It's expected to be in the "mini-plugin_<version>[-<prerelease>][+<metadata>]" format
func BinaryName(version *version.Version) string {
	retStr := fmt.Sprintf("mini-plugin_%s", version.Core())
	if version.Prerelease() != "" {
		retStr = fmt.Sprintf("%s-%s", retStr, version.Prerelease())
	}
	if version.Metadata() != "" {
		retStr = fmt.Sprintf("%s+%s", retStr, version.Metadata())
	}

	return retStr
}

// ExpectedInstalledName is the expected full name of the plugin once installed.
func ExpectedInstalledName(versionStr string) string {
	version.Must(version.NewVersion(versionStr))

	versionStr = strings.ReplaceAll(versionStr, "v", "")

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	return fmt.Sprintf("packer-plugin-tester_v%s_x%s.%s_%s_%s%s",
		versionStr,
		plugin.APIVersionMajor,
		plugin.APIVersionMinor,
		runtime.GOOS, runtime.GOARCH, ext)
}

// BuildSimplePlugin creates a plugin that essentially does nothing.
//
// The plugin's code is contained in a subdirectory of this, and lets us
// change the attributes of the plugin binary itself, like the SDK version,
// the plugin's version, etc.
//
// The plugin is functional, and can be used to run builds with.
// There won't be anything substantial created though, its goal is only
// to validate the core functionality of Packer.
//
// The path to the plugin is returned, it won't be removed automatically
// though, deletion is the caller's responsibility.
func (ts *PackerTestSuite) BuildSimplePlugin(versionString string, t *testing.T) string {
	// Only build plugin binary if not already done beforehand
	path, ok := ts.compiledPlugins.Load(versionString)
	if ok {
		return path.(string)
	}

	v := version.Must(version.NewSemver(versionString))

	t.Logf("Building plugin in version %v", v)

	testDir, err := currentDir()
	if err != nil {
		t.Fatalf("failed to compile plugin binary: %s", err)
	}

	testerPluginDir := filepath.Join(testDir, "plugin_tester")
	outBin := filepath.Join(ts.pluginsDirectory, BinaryName(v))

	compileCommand := exec.Command("go", "build", "-C", testerPluginDir, "-o", outBin, "-ldflags", LDFlags(v), ".")
	logs, err := compileCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to compile plugin binary: %s\ncompiler logs: %s", err, logs)
	}

	ts.compiledPlugins.Store(v.String(), outBin)

	return outBin
}

// MakePluginDir installs a list of plugins into a temporary directory and returns its path
//
// This can be set in the environment for a test through a function like t.SetEnv(), so
// packer will be able to use that directory for running its functions.
//
// Deletion of the directory is the caller's responsibility.
func (ts *PackerTestSuite) MakePluginDir(pluginVersions ...string) (pluginTempDir string, cleanup func()) {
	t := ts.T()

	for _, ver := range pluginVersions {
		ts.BuildSimplePlugin(ver, t)
	}

	var err error

	defer func() {
		if err != nil {
			if pluginTempDir != "" {
				os.RemoveAll(pluginTempDir)
			}
			t.Fatalf("failed to prepare temporary plugin directory %q: %s", pluginTempDir, err)
		}
	}()

	pluginTempDir, err = os.MkdirTemp("", "packer-plugin-dir-temp-")
	if err != nil {
		return
	}

	for _, pluginVersion := range pluginVersions {
		path, ok := ts.compiledPlugins.Load(pluginVersion)
		if !ok {
			err = fmt.Errorf("failed to get path to version %q, was it compiled?", pluginVersion)
		}
		cmd := ts.PackerCommand().SetArgs("plugins", "install", "--path", path.(string), "github.com/hashicorp/tester").AddEnv("PACKER_PLUGIN_PATH", pluginTempDir)
		cmd.Assert(MustSucceed())
		out, stderr, cmdErr := cmd.Run()
		if cmdErr != nil {
			err = fmt.Errorf("failed to install tester plugin version %q: %s\nCommand stdout: %s\nCommand stderr: %s", pluginVersion, err, out, stderr)
			return
		}
	}

	return pluginTempDir, func() {
		err := os.RemoveAll(pluginTempDir)
		if err != nil {
			t.Logf("failed to remove temporary plugin directory %q: %s. This may need manual intervention.", pluginTempDir, err)
		}
	}
}

// ManualPluginInstall emulates how Packer installs plugins with `packer plugins install`
//
// This is used for some tests if we want to install a plugin that cannot be installed
// through the normal commands (typically because Packer rejects it).
func ManualPluginInstall(t *testing.T, dest, srcPlugin, versionStr string) {
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		t.Fatalf("failed to create destination directories %q: %s", dest, err)
	}

	pluginName := ExpectedInstalledName(versionStr)
	destPath := filepath.Join(dest, pluginName)

	CopyFile(t, destPath, srcPlugin)

	shaPath := fmt.Sprintf("%s_SHA256SUM", destPath)
	WriteFile(t, shaPath, SHA256Sum(t, destPath))
}
