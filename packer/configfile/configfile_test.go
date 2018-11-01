package configfile

import (
	"os"
	"path/filepath"
	"io/ioutil"
	"fmt"
        "testing"
)

func testConfigTmpDir_impl(t *testing.T) string {
	var dir string

	prefix, _ := ConfigTmpDir()
	if dir, err := ioutil.TempDir(prefix, ""); err == nil {
		defer os.RemoveAll(dir)
	} else {
		_ := fmt.Errorf("Error making directory: %s", err)
	}

	return dir
}

func TestConfigTmpDir(t *testing.T) {
	testConfigTmpDir_impl(t)
}

func TestConfigTmpDir_noenv_PackerTmpDir(t *testing.T) {
	oldenv := os.Getenv(EnvPackerTmpDir)
	defer os.Setenv(EnvPackerTmpDir, oldenv)
	os.Setenv(EnvPackerTmpDir, "")

	dir1 := testConfigTmpDir_impl(t)

	cd, err := ConfigDir()
	if err != nil {
		t.Fatalf("bad ConfigDir")
	}
	td := filepath.Join(cd, "tmp")
	os.Setenv(EnvPackerTmpDir, td)

	dir2 := testConfigTmpDir_impl(t)

	if filepath.Dir(dir1) != filepath.Dir(dir2) {
		t.Fatalf("base directories do not match: %s %s", filepath.Dir(dir1), filepath.Dir(dir2))
	}
}

func TestConfigTmpDir_PackerTmpDir(t *testing.T) {
	oldenv := os.Getenv(EnvPackerTmpDir)
	defer os.Setenv(EnvPackerTmpDir, oldenv)
	os.Setenv(EnvPackerTmpDir, ".")

	dir1 := testConfigTmpDir_impl(t)

	abspath, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("bad absolute path")
	}
	dir2 := filepath.Join(abspath, "tmp")

	if filepath.Dir(dir1) != filepath.Dir(dir2) {
		t.Fatalf("base directories do not match: %s %s", filepath.Dir(dir1), filepath.Dir(dir2))
	}
}
