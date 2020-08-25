package command

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_hcl2_upgrade(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	_ = cwd

	tc := []struct {
		folder string
	}{
		{"hcl2_upgrade_basic"},
	}

	for _, tc := range tc {
		t.Run(tc.folder, func(t *testing.T) {
			inputPath := filepath.Join(testFixture(tc.folder, "input.json"))
			outputPath := inputPath + ".pkr.hcl"
			expectedPath := filepath.Join(testFixture(tc.folder, "expected.pkr.hcl"))
			p := helperCommand(t, "hcl2_upgrade", inputPath)
			bs, err := p.CombinedOutput()
			if err != nil {
				t.Fatalf("%v %s", err, bs)
			}
			expected := mustBytes(ioutil.ReadFile(expectedPath))
			actual := mustBytes(ioutil.ReadFile(outputPath))

			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Fatalf("unexpected output: %s", diff)
			}
			os.Remove(outputPath)
		})
	}
}

func mustBytes(b []byte, e error) []byte {
	if e != nil {
		panic(e)
	}
	return b
}
