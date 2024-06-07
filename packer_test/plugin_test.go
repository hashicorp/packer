package packer_test

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

var compiledPlugins = struct {
	pluginVersions map[string]string
	RWMutex        sync.RWMutex
}{
	pluginVersions: map[string]string{},
}

func StorePluginVersion(pluginVersion, path string) {
	compiledPlugins.RWMutex.Lock()
	defer compiledPlugins.RWMutex.Unlock()
	compiledPlugins.pluginVersions[pluginVersion] = path
}

func LoadPluginVersion(pluginVersion string) (string, bool) {
	compiledPlugins.RWMutex.RLock()
	defer compiledPlugins.RWMutex.RUnlock()

	path, ok := compiledPlugins.pluginVersions[pluginVersion]
	return path, ok
}

var tempPluginBinaryPath = struct {
	path string
	once sync.Once
}{}

// PluginBinaryDir returns the path to the directory where temporary binaries will be compiled
func PluginBinaryDir() string {
	tempPluginBinaryPath.once.Do(func() {
		tempDir, err := os.MkdirTemp("", "packer-core-acc-test-")
		if err != nil {
			panic(fmt.Sprintf("failed to create temporary directory for compiled plugins: %s", err))
		}

		tempPluginBinaryPath.path = tempDir
	})

	return tempPluginBinaryPath.path
}

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
func BuildSimplePlugin(versionString string, t *testing.T) string {
	// Only build plugin binary if not already done beforehand
	path, ok := LoadPluginVersion(versionString)
	if ok {
		return path
	}

	v := version.Must(version.NewSemver(versionString))

	t.Logf("Building plugin in version %v", v)

	testDir, err := currentDir()
	if err != nil {
		t.Fatalf("failed to compile plugin binary: %s", err)
	}

	testerPluginDir := filepath.Join(testDir, "plugin_tester")
	outBin := filepath.Join(PluginBinaryDir(), BinaryName(v))

	compileCommand := exec.Command("go", "build", "-C", testerPluginDir, "-o", outBin, "-ldflags", LDFlags(v), ".")
	logs, err := compileCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to compile plugin binary: %s\ncompiler logs: %s", err, logs)
	}

	StorePluginVersion(v.String(), outBin)

	return outBin
}

// currentDir returns the directory in which the current file is located.
//
// Since we're in tests it's reliable as they're supposed to run on the same
// machine the binary's compiled from, but goes to say it's not meant for use
// in distributed binaries.
func currentDir() (string, error) {
	// pc uintptr, file string, line int, ok bool
	_, testDir, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("couldn't get the location of the test suite file")
	}

	return filepath.Dir(testDir), nil
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
		BuildSimplePlugin(ver, t)
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
		path, ok := LoadPluginVersion(pluginVersion)
		if !ok {
			err = fmt.Errorf("failed to get path to version %q, was it compiled?", pluginVersion)
		}
		cmd := ts.PackerCommand().SetArgs("plugins", "install", "--path", path, "github.com/hashicorp/tester").AddEnv("PACKER_PLUGIN_PATH", pluginTempDir)
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

// CopyFile essentially replicates the `cp` command, for a file only.
//
// # Permissions are copied over from the source to destination
//
// The function detects if destination is a directory or a file (existent or not).
//
// If this is the former, we append the source file's basename to the
// directory and create the file from that inferred path.
func CopyFile(t *testing.T, dest, src string) {
	st, err := os.Stat(src)
	if err != nil {
		t.Fatalf("failed to stat origin file %q: %s", src, err)
	}

	// If the stat call fails, we assume dest is the destination file.
	dstStat, err := os.Stat(dest)
	if err == nil && dstStat.IsDir() {
		dest = filepath.Join(dest, filepath.Base(src))
	}

	destFD, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, st.Mode().Perm())
	if err != nil {
		t.Fatalf("failed to create cp destination file %q: %s", dest, err)
	}
	defer destFD.Close()

	srcFD, err := os.Open(src)
	if err != nil {
		t.Fatalf("failed to open source file to copy: %s", err)
	}
	defer srcFD.Close()

	_, err = io.Copy(destFD, srcFD)
	if err != nil {
		t.Fatalf("failed to copy from %q -> %q: %s", src, dest, err)
	}
}

// WriteFile writes `content` to a file `dest`
//
// The default permissions of that file is 0644
func WriteFile(t *testing.T, dest string, content string) {
	outFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("failed to open/create %q: %s", dest, err)
	}
	defer outFile.Close()

	_, err = fmt.Fprintf(outFile, content)
	if err != nil {
		t.Fatalf("failed to write to file %q: %s", dest, err)
	}
}

// TempWorkdir creates a working directory for a Packer test with the list of files
// given as input.
//
// The files should either have a path relative to the test that invokes it, or should
// be absolute.
// Each file will be copied to the root of the workdir being created.
//
// If any file cannot be found, this function will fail
func TempWorkdir(t *testing.T, files ...string) (string, func()) {
	var err error
	tempDir, err := os.MkdirTemp("", "packer-test-workdir-")
	if err != nil {
		t.Fatalf("failed to create temporary working directory: %s", err)
	}

	defer func() {
		if err != nil {
			os.RemoveAll(tempDir)
			t.Errorf("failed to create temporary workdir: %s", err)
		}
	}()

	for _, file := range files {
		CopyFile(t, tempDir, file)
	}

	return tempDir, func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Logf("failed to remove temporary workdir %q: %s. This will need manual action.", tempDir, err)
		}
	}
}

// SHA256Sum computes the SHA256 digest for an input file
//
// The digest is returned as a hexstring
func SHA256Sum(t *testing.T, file string) string {
	fl, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to compute sha256sum for %q: %s", file, err)
	}
	sha := sha256.New()
	sha.Write(fl)
	return fmt.Sprintf("%x", sha.Sum([]byte{}))
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
