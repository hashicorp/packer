package main

import (
	"fmt"

	"github.com/hashicorp/packer/packer_test/common/check"
)

type malformedDepTestCase struct {
	name          string
	command       string
	templatePath  string
	useSequential bool
}

func genMalformedDepTestCases() []malformedDepTestCase {
	retVals := []malformedDepTestCase{}

	for _, cmd := range []string{"build", "validate"} {
		for _, template := range []string{"./templates/malformed_data_dep.pkr.hcl", "./templates/malformed_local_dep.pkr.hcl"} {
			for _, seq := range []bool{true, false} {
				retVals = append(retVals, malformedDepTestCase{
					name: fmt.Sprintf("Malformed dep packer %s --use-sequential-evaluation=%t %s",
						cmd, seq, template),
					command:       cmd,
					templatePath:  template,
					useSequential: seq,
				})
			}
		}
	}

	return retVals
}

func (ts *PackerDAGTestSuite) TestMalformedDependency() {
	pluginDir := ts.MakePluginDir()
	defer pluginDir.Cleanup()

	for _, tc := range genMalformedDepTestCases() {
		ts.Run(tc.name, func() {
			ts.PackerCommand().UsePluginDir(pluginDir).
				SetArgs(tc.command,
					fmt.Sprintf("--use-sequential-evaluation=%t", tc.useSequential),
					tc.templatePath).
				Assert(check.MustFail())
		})
	}
}
