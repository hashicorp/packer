package hyperone

import (
	"os"
	"testing"

	builderT "github.com/hashicorp/packer/helper/builder/testing"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("HYPERONE_TOKEN"); v == "" {
		t.Fatal("HYPERONE_TOKEN must be set for acceptance tests")
	}
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"vm_type": "a1.nano",
		"source_image": "5c3fef425ceadc3d412beee5",
		"disk_size": 10
	}]
}
`
