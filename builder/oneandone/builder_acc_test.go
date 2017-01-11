package oneandone

import (
	"os"
	"testing"

	builderT "github.com/mitchellh/packer/helper/builder/testing"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ONEANDONE_TOKEN"); v == "" {
		t.Fatal("ONEANDONE_TOKEN must be set for acceptance tests")
	}
}

const testBuilderAccBasic = `
{
      "builders": [{
	      "type": "oneandone",
	      "disk_size": "50",
	      "snapshot_name": "test5",
	      "image" : "ubuntu1604-64min"
    }]
}
`
