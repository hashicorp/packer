package configfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func testConfigTmpDir_impl(t *testing.T) string {
	var dir string
	var err error

	prefix, _ := ConfigTmpDir()
	if dir, err = ioutil.TempDir(prefix, ""); err == nil {
		defer os.RemoveAll(dir)
	} else {
		t.Fatalf("Error making directory: %s", err)
	}

	return dir
}

func TestConfigTmpDir_noenv_PackerTmpDir(t *testing.T) {
	var oldenv, cd, dir1, dir2 string
	var err error

	oldenv = os.Getenv(EnvPackerTmpDir)
	defer os.Setenv(EnvPackerTmpDir, oldenv)
	os.Setenv(EnvPackerTmpDir, "")

	dir1 = testConfigTmpDir_impl(t)

	if cd, err = ConfigDir(); err != nil {
		t.Fatalf("Error during ConfigDir()")
	}

	os.Setenv(EnvPackerTmpDir, filepath.Join(cd, "tmp"))

	dir2 = testConfigTmpDir_impl(t)

	if filepath.Dir(dir1) != filepath.Dir(dir2) {
		t.Fatalf("base directories do not match: '%s' vs '%s'", filepath.Dir(dir1), filepath.Dir(dir2))
	}
}

func TestConfigTmpDir_PackerTmpDir(t *testing.T) {
	var oldenv, abspath, dir1, dir2 string
	var err error

	oldenv = os.Getenv(EnvPackerTmpDir)
	defer os.Setenv(EnvPackerTmpDir, oldenv)
	os.Setenv(EnvPackerTmpDir, ".")

	dir1 = testConfigTmpDir_impl(t)

	if abspath, err = filepath.Abs("."); err != nil {
		t.Fatalf("Error during filepath.ABS()")
	}

	dir2 = filepath.Join(abspath, "tmp")

	if filepath.Dir(dir1) != filepath.Dir(dir2) {
		t.Fatalf("Base directories do not match: '%s' vs '%s'", filepath.Dir(dir1), filepath.Dir(dir2))
	}
}
