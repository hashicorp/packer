package main

import (
	"fmt"

	"github.com/hashicorp/packer/packer_test/lib"
)

func (ts *PackerDAGTestSuite) TestWithBothDataLocalMixedOrder() {
	pluginDir, cleanup := ts.MakePluginDir()
	defer cleanup()

	for _, cmd := range []string{"build", "validate"} {
		ts.Run(fmt.Sprintf("%s: evaluating with DAG - success expected", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs(cmd, "./templates/mixed_data_local.pkr.hcl").
				Assert(lib.MustSucceed())
		})

		ts.Run(fmt.Sprintf("%s: evaluating sequentially - failure expected", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs(cmd, "--use-sequential-evaluation", "./templates/mixed_data_local.pkr.hcl").
				Assert(lib.MustFail())
		})
	}
}
