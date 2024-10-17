package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/packer/packer_test/common/check"
)

var outputFileRegexp = regexp.MustCompile("data/file - Writing to \"([^\"]+)\"")

func cleanupOutputFile(stderr string) error {
	matches := outputFileRegexp.FindStringSubmatch(stderr)
	if len(matches) != 2 {
		return fmt.Errorf("cannot match file datasource from packer output")
	}

	filePath := matches[1]
	return os.Remove(filePath)
}

// TestWithNothing checks that in its simplest form, the datasource succeeds and writes an empty file to TMPDIR
func (ts *FileDatasourceTestSuite) TestWithNothing() {
	pd := ts.MakePluginDir()
	defer pd.Cleanup()

	cmd := ts.PackerCommand().UsePluginDir(pd).
		SetArgs("build", "./templates/file_simplest.pkr.hcl")
	cmd.Assert(check.MustSucceed(),
		check.Grep("data/file - Writing to", check.GrepStderr))

	_, stderr, _ := cmd.Output()
	err := cleanupOutputFile(stderr)
	if err != nil {
		ts.T().Logf("failed to find file to cleanup from stderr, will need some manual action")
	}
}

// TestWithContents checks that the datasource writes what is expected to the output file, in TMPDIR
func (ts *FileDatasourceTestSuite) TestWithContents() {
	pd := ts.MakePluginDir()
	defer pd.Cleanup()

	cmd := ts.PackerCommand().UsePluginDir(pd).
		SetArgs("build", "./templates/file_with_contents.pkr.hcl")
	cmd.Assert(check.MustSucceed(),
		check.Grep("data/file - Writing to", check.GrepStderr),
		check.Grep("file contents: Hello there!"))

	_, stderr, _ := cmd.Output()
	err := cleanupOutputFile(stderr)
	if err != nil {
		ts.T().Logf("failed to find file to cleanup from stderr, will need some manual action")
	}
}

// TestWithFileDestination checks that we can specify a file directory, with its hierarchy existing in the first place
func (ts *FileDatasourceTestSuite) TestWithFileDestination() {
	pd := ts.MakePluginDir()
	defer pd.Cleanup()

	// Create full hierarchy for output directory
	err := os.MkdirAll("out_dir/subdir", 0755)
	if err != nil {
		ts.T().Fatalf("failed to create output directory: %s", err)
	}
	defer os.RemoveAll("out_dir")

	// No need to clean output file, since the directory is cleaned-up automatically
	ts.PackerCommand().UsePluginDir(pd).
		SetArgs("build", "./templates/local_destination.pkr.hcl").
		Assert(check.MustSucceed(),
			check.Grep("data/file - Writing to", check.GrepStderr),
			check.Grep("file contents: Hello there!"),
			check.FileExists("out_dir/subdir/out.txt", false))
}

// TestWithFileDestinationAlreadyExists checks that we can specify a file output, even if it exists, and the output is strictly the contents of the file
func (ts *FileDatasourceTestSuite) TestWithFileDestinationAlreadyExists() {
	pd := ts.MakePluginDir()
	defer pd.Cleanup()

	// Create full hierarchy for output directory
	err := os.MkdirAll("out_dir/subdir", 0755)
	if err != nil {
		ts.T().Fatalf("failed to create output directory: %s", err)
	}
	defer os.RemoveAll("out_dir")

	err = os.WriteFile("out_dir/subdir/out.txt", []byte("Hello there!\n"), 0644)
	if err != nil {
		ts.T().Fatalf("failed to write output file 'out_dir/subdir/out.txt' before test: %s", err)
	}

	// No need to clean output file, since the directory is cleaned-up automatically
	ts.PackerCommand().UsePluginDir(pd).
		SetArgs("build", "./templates/local_destination.pkr.hcl").
		Assert(check.MustSucceed(),
			check.Grep("data/file - Writing to", check.GrepStderr),
			check.Grep("file contents: Hello there!"),
			check.MkPipeCheck("only one occurrence in contents of output file",
				check.PipeGrep("Hello there!"), check.LineCount()).
				SetStream(check.OnlyStdout).
				SetTester(check.IntCompare(check.Eq, 1)),
			check.FileExists("out_dir/subdir/out.txt", false))
}

// TestWithFileDestination checks that we can specify a destination directory, with it existing in the first place
func (ts *FileDatasourceTestSuite) TestWithDirectoryDestination() {
	pd := ts.MakePluginDir()
	defer pd.Cleanup()

	err := os.MkdirAll("out_dir/subdir", 0755)
	if err != nil {
		ts.T().Fatalf("failed to create output directory: %s", err)
	}
	// Cleanup output directory
	defer os.RemoveAll("out_dir")

	ts.PackerCommand().UsePluginDir(pd).
		SetArgs("build", "./templates/local_dir_destination.pkr.hcl").
		Assert(check.MustSucceed(),
			check.Grep("data/file - Writing to", check.GrepStderr),
			check.Grep("file contents: Hello there!"),
			check.FileExists("out_dir/subdir", true))
}

// TestWithFileDestinationNoPreCreate checks that we can specify a destination directory, without it existing in the first place
func (ts *FileDatasourceTestSuite) TestWithDirectoryDestinationNoPreCreate() {
	pd := ts.MakePluginDir()
	defer pd.Cleanup()

	ts.PackerCommand().UsePluginDir(pd).
		SetArgs("build", "./templates/local_dir_destination.pkr.hcl").
		Assert(check.MustSucceed(),
			check.Grep("data/file - Writing to", check.GrepStderr),
			check.Grep("file contents: Hello there!"),
			check.FileExists("out_dir/subdir", true))

	// Cleanup output directory
	os.RemoveAll("out_dir")
}

// TestWithTempDirNotWritable checks that the datasource fails if the temporary directory is not writable, and we did not provide a Destination.
//
// NOTE: this one fails to execute completely since Packer needs TMPDIR to be writable for logs, and changing this may include more work.
// Leaving it here still if that changes, to be sure we don't have an unexpected crash if that changes.
func (ts *FileDatasourceTestSuite) TestWithTempDirNotWritable() {
	pd := ts.MakePluginDir()
	defer pd.Cleanup()

	tempName := "fake_temp"
	err := os.Mkdir(tempName, 0555)
	if err != nil {
		ts.T().Fatalf("failed to create temporary tmpdir %q: %s", tempName, err)
	}
	defer os.RemoveAll(tempName)

	ts.PackerCommand().UsePluginDir(pd).
		SetArgs("build", "./templates/file_with_contents.pkr.hcl").
		AddEnv("TMPDIR", tempName).
		Assert(check.MustFail())
}

// TestWithDestDirNotWritable checks that the datasource fails if the destination directory is not writable, and a destination is provided.
func (ts *FileDatasourceTestSuite) TestWithDestDirNotWritable() {
	pd := ts.MakePluginDir()
	defer pd.Cleanup()

	err := os.MkdirAll("out_dir", 0555)
	if err != nil {
		ts.T().Fatalf("failed to create output directory: %s", err)
	}
	defer func() {
		err := os.RemoveAll("out_dir")
		if err != nil {
			ts.T().Logf("failed to remove out_dir: %s", err)
		}
	}()

	ts.PackerCommand().UsePluginDir(pd).
		SetArgs("build", "./templates/local_destination.pkr.hcl").
		Assert(check.MustFail())
}
