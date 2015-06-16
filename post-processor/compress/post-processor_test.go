package compress

// import (
// 	"testing"
//
// 	builderT "github.com/mitchellh/packer/helper/builder/testing"
// )
//
// func TestBuilderTagsAcc_basic(t *testing.T) {
// 	builderT.Test(t, builderT.TestCase{
// 		Builder:  &Builder{},
// 		Template: simpleTestCase,
// 		Check:    checkTags(),
// 	})
// }

const simpleTestCase = `
{
  "type": "compress",
  "output": "foo.tar.gz"
}
`
