package profitbricks

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
	if v := os.Getenv("PROFITBRICKS_USERNAME"); v == "" {
		t.Fatal("PROFITBRICKS_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("PROFITBRICKS_PASSWORD"); v == "" {
		t.Fatal("PROFITBRICKS_PASSWORD must be set for acceptance tests")
	}
}

const testBuilderAccBasic = `
{
	"builders": [{
	      "image": "Ubuntu-16.04",
	      "password": "password",
	      "username": "username",
	      "snapshot_name": "packer",
	      "type": "profitbricks"
   	}]
}
`
