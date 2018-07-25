package vagrant

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVBoxProvider_impl(t *testing.T) {
	var _ Provider = new(VBoxProvider)
}

func TestDecomressOVA(t *testing.T) {
	td, err := ioutil.TempDir("", "pp-vagrant-virtualbox")
	assert.NoError(t, err)
	fixture := "../../common/test-fixtures/decompress-tar/outside_parent.tar"
	err = DecompressOva(td, fixture)
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(filepath.Base(td), "demo.poc"))
	assert.Error(t, err)
	_, err = os.Stat(filepath.Join(td, "demo.poc"))
	assert.NoError(t, err)
	os.RemoveAll(td)
}
