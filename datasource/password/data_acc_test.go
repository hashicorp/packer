// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package password

import (
	_ "embed"
	"fmt"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/basic.pkr.hcl
var testDatasourceBasic string

//go:embed test-fixtures/basic-custom-charset.pkr.hcl
var testDatasourceBasicWithCustomCharset string

//go:embed test-fixtures/invalid-length.pkr.hcl
var testDataSourceInvalidLength string

//go:embed test-fixtures/empty-charset.pkr.hcl
var testDataSourceEmptyCharset string

func TestPasswordDataSource(t *testing.T) {
	tests := []struct {
		Name  string
		Path  string
		Error bool
	}{
		{
			Name:  "basic_test",
			Path:  testDatasourceBasic,
			Error: false,
		},
		{
			Name:  "basic_with_custom_charset_test",
			Path:  testDatasourceBasicWithCustomCharset,
			Error: false,
		},
		{
			Name:  "invalid_length_test",
			Path:  testDataSourceInvalidLength,
			Error: true,
		},
		{
			Name:  "empty_charset_test",
			Path:  testDataSourceEmptyCharset,
			Error: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testCase := &acctest.PluginTestCase{
				Name:     tt.Name,
				Template: tt.Path,
				Check: func(buildCommand *exec.Cmd, logfile string) error {
					if buildCommand.ProcessState != nil {
						if buildCommand.ProcessState.ExitCode() != 0 && !tt.Error {
							return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
						}
						if tt.Error && buildCommand.ProcessState.ExitCode() == 0 {
							return fmt.Errorf("Expected Bad exit code.")
						}
					}
					return nil
				},
			}
			acctest.TestPlugin(t, testCase)
		})
	}

}
