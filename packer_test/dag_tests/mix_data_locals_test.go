package main

import (
	"fmt"

	"github.com/hashicorp/packer/packer_test/common/check"
)

func (ts *PackerDAGTestSuite) TestWithBothDataLocalMixedOrder() {
	pluginDir := ts.MakePluginDir()
	defer pluginDir.Cleanup()

	for _, cmd := range []string{"build", "validate"} {
		ts.Run(fmt.Sprintf("%s: evaluating with DAG - success expected", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs(cmd, "./templates/mixed_data_local.pkr.hcl").
				Assert(check.MustSucceed())
		})

		ts.Run(fmt.Sprintf("%s: evaluating sequentially - failure expected", cmd), func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs(cmd, "--use-sequential-evaluation", "./templates/mixed_data_local.pkr.hcl").
				Assert(check.MustFail())
		})
	}
}
