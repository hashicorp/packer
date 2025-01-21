package plugin_tests

import (
	"os"

	"github.com/hashicorp/packer/packer_test/common/check"
)

func (ts *PackerHCPSbomTestSuite) TestSourceNotExisting() {
	ts.SkipNoAcc()

	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "templates/source_not_existing.pkr.hcl").
		Assert(check.MustFail(), check.Grep("download failed for SBOM file"))
}

// Greayed out because the communicator for the docker plugin does not return an error
// when downloading a full directory, instead it returns a 0-byte stream without an error.
//
// So the sbom provisioner fails with a validation error instead of a file not found type
// of error.
//
// func (ts *PackerHCPSbomTestSuite) TestSourceIsDir() {
// 	ts.SkipNoAcc()
//
// 	path, cleanup := ts.MakePluginDir()
// 	defer cleanup()
//
// 	ts.PackerCommand().UsePluginDir(path).
// 		SetArgs("plugins", "install", "github.com/hashicorp/docker").
// 		Assert(check.MustSucceed())
//
// 	ts.PackerCommand().UsePluginDir(path).
// 		SetArgs("build", "templates/source_is_dir.pkr.hcl").
// 		Assert(check.MustFail(), check.Grep("download failed for SBOM file"), check.Dump(ts.T()))
// }

// * output file - does not exist, and intermediate dirs don't exist
func (ts *PackerHCPSbomTestSuite) TestDestFile_NoIntermediateDirs() {
	ts.SkipNoAcc()

	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "./templates/dest_is_file_no_interm_dirs.pkr.hcl").
		Assert(check.MustSucceed(), check.FileExists("sbom/sbom_cyclonedx.json", false))

	os.RemoveAll("sbom")
}

// * output file - does not exist, and intermediate dirs already exist
func (ts *PackerHCPSbomTestSuite) TestDestFile_WithIntermediateDirs() {
	ts.SkipNoAcc()

	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	os.MkdirAll("sbom", 0755)

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "./templates/dest_is_file_no_interm_dirs.pkr.hcl").
		Assert(check.MustSucceed(), check.FileExists("sbom/sbom_cyclonedx.json", false))

	os.RemoveAll("sbom")
}

// * output directory (without trailing slash) - directory exists
func (ts *PackerHCPSbomTestSuite) TestDestDir_NoTrailingSlash() {
	ts.SkipNoAcc()

	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	os.MkdirAll("sbom", 0755)

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "./templates/dest_is_dir.pkr.hcl").
		Assert(check.MustSucceed(), check.FileGlob("./sbom/packer-user-sbom-*.json"))

	os.RemoveAll("sbom")
}

// * output directory (with trailing slash) - directory exists
func (ts *PackerHCPSbomTestSuite) TestDestDir_WithTrailingSlash() {
	ts.SkipNoAcc()

	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	os.MkdirAll("sbom", 0755)

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "./templates/dest_is_dir_with_trailing_slash.pkr.hcl").
		Assert(check.MustSucceed(), check.FileGlob("./sbom/packer-user-sbom-*.json"))

	os.RemoveAll("sbom")
}

// * output directory (with trailing slash) - directory doesn't exist
func (ts *PackerHCPSbomTestSuite) TestDestDir_WithTrailingSlash_NoDir() {
	ts.SkipNoAcc()

	dir := ts.MakePluginDir()
	defer dir.Cleanup()

	ts.PackerCommand().UsePluginDir(dir).
		SetArgs("plugins", "install", "github.com/hashicorp/docker").
		Assert(check.MustSucceed())

	ts.PackerCommand().UsePluginDir(dir).
		AddEnv("HOME", os.Getenv("HOME")).
		AddEnv("PATH", os.Getenv("PATH")).
		SetArgs("build", "./templates/dest_is_dir_with_trailing_slash.pkr.hcl").
		Assert(check.MustSucceed(), check.FileGlob("./sbom/packer-user-sbom-*.json"))

	os.RemoveAll("sbom")
}
