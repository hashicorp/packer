package common

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer/packer_test/common/check"
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

// BuildCustomisation is a function that allows you to change things on a plugin's
// local files, with a way to rollback those changes after the fact.
//
// The function is meant to take a path parameter to the directory for the plugin,
// and returns a function that unravels those changes once the build process is done.
type BuildCustomisation func(string) (error, func())

const SDKModule = "github.com/hashicorp/packer-plugin-sdk"

// UseDependency invokes go get and go mod tidy to update a package required
// by the plugin, and use it to build the plugin with that change.
func UseDependency(remoteModule, ref string) BuildCustomisation {
	return func(path string) (error, func()) {
		modPath := filepath.Join(path, "go.mod")

		stat, err := os.Stat(modPath)
		if err != nil {
			return fmt.Errorf("cannot stat mod file %q: %s", modPath, err), nil
		}

		// Save old go.mod file from dir
		oldGoMod, err := os.ReadFile(modPath)
		if err != nil {
			return fmt.Errorf("failed to read current mod file %q: %s", modPath, err), nil
		}

		modSpec := fmt.Sprintf("%s@%s", remoteModule, ref)
		cmd := exec.Command("go", "get", modSpec)
		cmd.Dir = path
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to run go get %s: %s", modSpec, err), nil
		}

		cmd = exec.Command("go", "mod", "tidy")
		cmd.Dir = path
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to run go mod tidy: %s", err), nil
		}

		return nil, func() {
			err = os.WriteFile(modPath, oldGoMod, stat.Mode())
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to reset modfile %q: %s; manual cleanup may be needed", modPath, err)
			}
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = path
			_ = cmd.Run()
		}
	}
}

// GetPluginPath gets the path for a pre-compiled plugin in the current test suite.
//
// The version only is needed, as the path to a compiled version of the tester
// plugin will be returned, so it can be installed after the fact.
//
// If the version requested does not exist, the function will panic.
func (ts *PackerTestSuite) GetPluginPath(t *testing.T, version string) string {
	path, ok := ts.compiledPlugins.Load(version)
	if !ok {
		t.Fatalf("tester plugin in version %q was not built, either build it during suite init, or with BuildTestPlugin", version)
	}

	return path.(string)
}

type CompilationResult struct {
	Error   error
	Version string
}

// Ready processes a series of CompilationResults, as returned by CompilePlugin
//
// If any of the jobs requested failed, the test will fail also.
func Ready(t *testing.T, results []chan CompilationResult) {
	for _, res := range results {
		jobErr := <-res
		empty := CompilationResult{}
		if jobErr != empty {
			t.Errorf("failed to compile plugin at version %s: %s", jobErr.Version, jobErr.Error)
		}
	}

	if t.Failed() {
		t.Fatalf("some plugins failed to be compiled, see logs for more info")
	}
}

type compilationJob struct {
	versionString  string
	suite          *PackerTestSuite
	done           bool
	resultCh       chan CompilationResult
	customisations []BuildCustomisation
}

// CompilationJobs keeps a queue of compilation jobs for plugins
//
// This approach allows us to avoid conflicts between compilation jobs.
// Typically building the plugin with different ldflags is safe to perform
// in parallel on the same file set, however customisations tend to be more
// conflictual, as two concurrent compilation jobs may end-up compiling the
// wrong plugin, which may cause some tests to misbehave, or even compilation
// jobs to fail.
//
// The solution to this approach is to have a global queue for every plugin
// compilation to be performed safely.
var CompilationJobs = make(chan compilationJob, 10)

// CompilePlugin builds a tester plugin with the specified version.
//
// The plugin's code is contained in a subdirectory of this file, and lets us
// change the attributes of the plugin binary itself, like the SDK version,
// the plugin's version, etc.
//
// The plugin is functional, and can be used to run builds with.
// There won't be anything substantial created though, its goal is only
// to validate the core functionality of Packer.
//
// The path to the plugin is returned, it won't be removed automatically
// though, deletion is the caller's responsibility.
//
// Note: each tester plugin may only be compiled once for a specific version in
// a test suite. The version may include core (mandatory), pre-release and
// metadata. Unlike Packer core, metadata does matter for the version being built.
//
// Note: the compilation will process asynchronously, and should be waited upon
// before tests that use this plugin may proceed. Refer to the `Ready` function
// for doing that.
func (ts *PackerTestSuite) CompilePlugin(versionString string, customisations ...BuildCustomisation) chan CompilationResult {
	resultCh := make(chan CompilationResult)

	CompilationJobs <- compilationJob{
		versionString:  versionString,
		suite:          ts,
		customisations: customisations,
		done:           false,
		resultCh:       resultCh,
	}

	return resultCh
}

func init() {
	// Run a processor coroutine for the duration of the test.
	//
	// It's simpler to have this occurring on the side at all times, without
	// trying to manage its lifecycle based on the current amount of queued
	// tasks, since this is linked to the test lifecycle, and as it's a single
	// coroutine, we can leave it run until the process exits.
	go func() {
		for job := range CompilationJobs {
			log.Printf("compiling plugin on version %s", job.versionString)
			err := compilePlugin(job.suite, job.versionString, job.customisations...)
			if err != nil {
				job.resultCh <- CompilationResult{
					Error:   err,
					Version: job.versionString,
				}
			}
			close(job.resultCh)
		}
	}()
}

// compilePlugin performs the actual compilation procedure for the plugin, and
// registers it to the test suite instance passed as a parameter.
func compilePlugin(ts *PackerTestSuite, versionString string, customisations ...BuildCustomisation) error {
	// Fail to build plugin if already built.
	//
	// Especially with customisations being a thing, relying on cache to get and
	// build a plugin at once means that the function is not idempotent anymore,
	// and therefore we cannot rely on it being called twice and producing the
	// same result, so we forbid it.
	if _, ok := ts.compiledPlugins.Load(versionString); ok {
		return fmt.Errorf("plugin version %q was already built, use GetTestPlugin instead", versionString)
	}

	v := version.Must(version.NewSemver(versionString))

	testDir, err := currentDir()
	if err != nil {
		return fmt.Errorf("failed to compile plugin binary: %s", err)
	}

	testerPluginDir := filepath.Join(testDir, "plugin_tester")
	for _, custom := range customisations {
		err, cleanup := custom(testerPluginDir)
		if err != nil {
			return fmt.Errorf("failed to prepare plugin workdir: %s", err)
		}
		defer cleanup()
	}

	outBin := filepath.Join(ts.pluginsDirectory, BinaryName(v))

	compileCommand := exec.Command("go", "build", "-C", testerPluginDir, "-o", outBin, "-ldflags", LDFlags(v), ".")
	logs, err := compileCommand.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to compile plugin binary: %s\ncompiler logs: %s", err, logs)
	}

	ts.compiledPlugins.Store(v.String(), outBin)
	return nil
}

type PluginDirSpec struct {
	dirPath string
	suite   *PackerTestSuite
}

// MakePluginDir installs a list of plugins into a temporary directory and returns its path
//
// This can be set in the environment for a test through a function like t.SetEnv(), so
// packer will be able to use that directory for running its functions.
//
// Deletion of the directory is the caller's responsibility.
//
// Note: all of the plugin versions specified to be installed in this plugin directory
// must have been compiled beforehand.
func (ts *PackerTestSuite) MakePluginDir() *PluginDirSpec {
	var err error

	pluginTempDir, err := os.MkdirTemp("", "packer-plugin-dir-temp-")
	if err != nil {
		return nil
	}

	return &PluginDirSpec{
		dirPath: pluginTempDir,
		suite:   ts,
	}
}

// InstallPluginVersions installs several versions of the tester plugin under
// github.com/hashicorp/tester.
//
// Each version of the plugin needs to have been pre-compiled.
//
// If a plugin is missing, the temporary directory will be removed.
func (ps *PluginDirSpec) InstallPluginVersions(pluginVersions ...string) *PluginDirSpec {
	t := ps.suite.T()

	var err error

	defer func() {
		if err != nil || t.Failed() {
			rmErr := os.RemoveAll(ps.Dir())
			if rmErr != nil {
				t.Logf("failed to remove temporary plugin directory %q: %s. This may need manual intervention.", ps.Dir(), err)
			}
			t.Fatalf("failed to install plugins to temporary plugin directory %q: %s", ps.Dir(), err)
		}
	}()

	for _, pluginVersion := range pluginVersions {
		path := ps.suite.GetPluginPath(t, pluginVersion)
		cmd := ps.suite.PackerCommand().SetArgs("plugins", "install", "--path", path, "github.com/hashicorp/tester").AddEnv("PACKER_PLUGIN_PATH", ps.Dir())
		cmd.Assert(check.MustSucceed())
		out, stderr, cmdErr := cmd.run()
		if cmdErr != nil {
			err = fmt.Errorf("failed to install tester plugin version %q: %s\nCommand stdout: %s\nCommand stderr: %s", pluginVersion, err, out, stderr)
		}
	}

	return ps
}

// Dir returns the temporary plugin dir for use in other functions
func (ps PluginDirSpec) Dir() string {
	return ps.dirPath
}

func (ps *PluginDirSpec) Cleanup() {
	pluginDir := ps.Dir()
	if pluginDir == "" {
		return
	}

	err := os.RemoveAll(pluginDir)
	if err != nil {
		ps.suite.T().Logf("failed to remove temporary plugin directory %q: %s. This may need manual intervention.", pluginDir, err)
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
