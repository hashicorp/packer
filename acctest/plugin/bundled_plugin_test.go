package plugin

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	amazonacc "github.com/hashicorp/packer-plugin-amazon/builder/ebs/acceptance"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

//go:embed test-fixtures/basic_amazon_bundled.pkr.hcl
var basicAmazonBundledEbsTemplate string

func TestAccBuildBundledPlugins(t *testing.T) {
	plugin := addrs.Plugin{
		Hostname:  "github.com",
		Namespace: "hashicorp",
		Type:      "amazon",
	}
	testCase := &acctest.PluginTestCase{
		Name: "amazon-ebs_bundled_test",
		Setup: func() error {
			return cleanupPluginInstallation(plugin)
		},
		Teardown: func() error {
			helper := amazonacc.AMIHelper{
				Region: "us-east-1",
				Name:   "packer-plugin-bundled-amazon-ebs-test",
			}
			return helper.CleanUpAmi()
		},
		Template: basicAmazonBundledEbsTemplate,
		Type:     "amazon-ebs",
		Init:     false,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			rawLogs, err := os.ReadFile(logfile)
			if err != nil {
				return fmt.Errorf("failed to read logs: %s", err)
			}

			var errs error

			logs := string(rawLogs)

			if !strings.Contains(logs, "Warning: Bundled plugins used") {
				errs = multierror.Append(errs, errors.New("expected warning about bundled plugins used, did not find it"))
			}

			if !strings.Contains(logs, "Then run 'packer init' to manage installation of the plugins") {
				errs = multierror.Append(errs, errors.New("expected suggestion about packer init in logs, did not find it."))
			}

			return errs
		},
	}

	acctest.TestPlugin(t, testCase)
}

//go:embed test-fixtures/basic_amazon_with_required_plugins.pkr.hcl
var basicAmazonRequiredPluginEbsTemplate string

func TestAccBuildBundledPluginsWithRequiredPlugins(t *testing.T) {
	plugin := addrs.Plugin{
		Hostname:  "github.com",
		Namespace: "hashicorp",
		Type:      "amazon",
	}
	testCase := &acctest.PluginTestCase{
		Name: "amazon-ebs_with_required_plugins_test",
		Setup: func() error {
			return cleanupPluginInstallation(plugin)
		},
		Teardown: func() error {
			helper := amazonacc.AMIHelper{
				Region: "us-east-1",
				Name:   "packer-plugin-required-plugin-amazon-ebs-test",
			}
			return helper.CleanUpAmi()
		},
		Template: basicAmazonRequiredPluginEbsTemplate,
		Type:     "amazon-ebs",
		Init:     false,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 1 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			rawLogs, err := os.ReadFile(logfile)
			if err != nil {
				return fmt.Errorf("failed to read logs: %s", err)
			}

			var errs error

			logs := string(rawLogs)

			if strings.Contains(logs, "Warning: Bundled plugins used") {
				errs = multierror.Append(errs, errors.New("did not expect warning about bundled plugins used"))
			}

			if !strings.Contains(logs, "Missing plugins") {
				errs = multierror.Append(errs, errors.New("expected error about plugins required and not installed, did not find it"))
			}

			return errs
		},
	}

	acctest.TestPlugin(t, testCase)
}

//go:embed test-fixtures/basic_amazon_bundled.json
var basicAmazonBundledEbsTemplateJSON string

func TestAccBuildBundledPluginsJSON(t *testing.T) {
	plugin := addrs.Plugin{
		Hostname:  "github.com",
		Namespace: "hashicorp",
		Type:      "amazon",
	}
	testCase := &acctest.PluginTestCase{
		Name: "amazon-ebs_bundled_test_json",
		Setup: func() error {
			return cleanupPluginInstallation(plugin)
		},
		Teardown: func() error {
			helper := amazonacc.AMIHelper{
				Region: "us-east-1",
				Name:   "packer-plugin-bundled-amazon-ebs-test-json",
			}
			return helper.CleanUpAmi()
		},
		Template: basicAmazonBundledEbsTemplateJSON,
		Type:     "amazon-ebs",
		Init:     false,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			rawLogs, err := os.ReadFile(logfile)
			if err != nil {
				return fmt.Errorf("failed to read logs: %s", err)
			}

			var errs error

			logs := string(rawLogs)

			if !strings.Contains(logs, "Warning: Bundled plugins used") {
				errs = multierror.Append(errs, errors.New("expected warning about bundled plugins, did not find it."))
			}

			if !strings.Contains(logs, "plugins with the 'packer plugins install' command") {
				errs = multierror.Append(errs, errors.New("expected suggestion about packer plugins install in logs, did not find it."))
			}

			return errs
		},
	}

	acctest.TestPlugin(t, testCase)
}
