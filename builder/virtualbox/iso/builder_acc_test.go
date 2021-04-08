package iso

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	builderT "github.com/hashicorp/packer/acctest"
)

func TestBuilderAcc_basic(t *testing.T) {
	templatePath := filepath.Join("testdata", "minimal.json")
	bytes, err := ioutil.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("failed to load template file %s", templatePath)
	}

	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: string(bytes),
	})
}
