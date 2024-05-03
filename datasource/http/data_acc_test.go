// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package http

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/basic.pkr.hcl
var testDatasourceBasic string

//go:embed test-fixtures/empty_url.pkr.hcl
var testDatasourceEmptyUrl string

//go:embed test-fixtures/404_url.pkr.hcl
var testDatasource404Url string

func TestHttpDataSource(t *testing.T) {
	tests := []struct {
		Name    string
		Path    string
		Error   bool
		Outputs map[string]string
	}{
		{
			Name:  "basic_test",
			Path:  testDatasourceBasic,
			Error: false,
			Outputs: map[string]string{
				"url": "url is https://www.packer.io/",
				// Check that body is not empty
				"body": "body is true",
			},
		},
		{
			Name:  "url_is_empty",
			Path:  testDatasourceEmptyUrl,
			Error: true,
			Outputs: map[string]string{
				"error": "the `url` must be specified",
			},
		},
		{
			Name:  "404_url",
			Path:  testDatasource404Url,
			Error: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testCase := &acctest.PluginTestCase{
				Name: tt.Name,
				Setup: func() error {
					return nil
				},
				Teardown: func() error {
					return nil
				},
				Template: tt.Path,
				Type:     "http",
				Check: func(buildCommand *exec.Cmd, logfile string) error {
					if buildCommand.ProcessState != nil {
						if buildCommand.ProcessState.ExitCode() != 0 && !tt.Error {
							return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
						}
						if tt.Error && buildCommand.ProcessState.ExitCode() == 0 {
							return fmt.Errorf("Expected Bad exit code.")
						}
					}

					if tt.Outputs != nil {
						logs, err := os.Open(logfile)
						if err != nil {
							return fmt.Errorf("Unable find %s", logfile)
						}
						defer logs.Close()

						logsBytes, err := io.ReadAll(logs)
						if err != nil {
							return fmt.Errorf("Unable to read %s", logfile)
						}
						logsString := string(logsBytes)

						for key, val := range tt.Outputs {
							if matched, _ := regexp.MatchString(val+".*", logsString); !matched {
								t.Fatalf(
									"logs doesn't contain expected log %v with value %v in %q",
									key,
									val,
									logsString)
							}
						}

					}

					return nil
				},
			}
			acctest.TestPlugin(t, testCase)
		})
	}

}
