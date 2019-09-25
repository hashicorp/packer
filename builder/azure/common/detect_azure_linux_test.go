package common

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAzure(t *testing.T) {
	f, err := ioutil.TempFile("", "TestIsAzure*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.Seek(0, 0)
	f.Truncate(0)
	f.Write([]byte("not the azure assettag"))

	assert.False(t, isAzureAssetTag(f.Name()), "asset tag is not Azure's")

	f.Seek(0, 0)
	f.Truncate(0)
	f.Write(azureAssetTag)

	assert.True(t, isAzureAssetTag(f.Name()), "asset tag is Azure's")
}
