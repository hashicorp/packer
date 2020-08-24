package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/hashicorp/packer/builder/amazon/ebs"
	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/builder/null"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/manifest"
	shell_local_pp "github.com/hashicorp/packer/post-processor/shell-local"
	filep "github.com/hashicorp/packer/provisioner/file"
	"github.com/hashicorp/packer/provisioner/shell"
	shell_local "github.com/hashicorp/packer/provisioner/shell-local"
	"github.com/hashicorp/packer/version"
)

// HasExec reports whether the current system can start new processes
// using os.StartProcess or (more commonly) exec.Command.
func HasExec() bool {
	switch runtime.GOOS {
	case "js":
		return false
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return false
		}
	}
	return true
}

// MustHaveExec checks that the current system can start new processes
// using os.StartProcess or (more commonly) exec.Command.
// If not, MustHaveExec calls t.Skip with an explanation.
func MustHaveExec(t testing.TB) {
	if !HasExec() {
		t.Skipf("skipping test: cannot exec subprocess on %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

func helperCommandContext(t *testing.T, ctx context.Context, s ...string) (cmd *exec.Cmd) {
	MustHaveExec(t)

	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	if ctx != nil {
		cmd = exec.CommandContext(ctx, os.Args[0], cs...)
	} else {
		cmd = exec.Command(os.Args[0], cs...)
	}
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func helperCommand(t *testing.T, s ...string) *exec.Cmd {
	return helperCommandContext(t, nil, s...)
}

// TestHelperProcess isn't a real test. It's used as a helper process
// for TestParameterRun.
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "console":
		os.Exit((&ConsoleCommand{Meta: commandMeta()}).Run(args))
	case "inspect":
		os.Exit((&InspectCommand{Meta: commandMeta()}).Run(args))
	case "build":
		os.Exit((&BuildCommand{Meta: commandMeta()}).Run(args))
	case "hcl2_upgrade":
		os.Exit((&HCL2UpgradeCommand{Meta: commandMeta()}).Run(args))
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}

func commandMeta() Meta {
	basicUi := &packer.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stdout,
	}

	CommandMeta := Meta{
		CoreConfig: &packer.CoreConfig{
			Components: getBareComponentFinder(),
			Version:    version.Version,
		},
		Ui: basicUi,
	}
	return CommandMeta
}

func getBareComponentFinder() packer.ComponentFinder {
	return packer.ComponentFinder{
		BuilderStore: packer.MapOfBuilder{
			"file":       func() (packer.Builder, error) { return &file.Builder{}, nil },
			"null":       func() (packer.Builder, error) { return &null.Builder{}, nil },
			"amazon-ebs": func() (packer.Builder, error) { return &ebs.Builder{}, nil },
		},
		ProvisionerStore: packer.MapOfProvisioner{
			"shell-local": func() (packer.Provisioner, error) { return &shell_local.Provisioner{}, nil },
			"shell":       func() (packer.Provisioner, error) { return &shell.Provisioner{}, nil },
			"file":        func() (packer.Provisioner, error) { return &filep.Provisioner{}, nil },
		},
		PostProcessorStore: packer.MapOfPostProcessor{
			"shell-local": func() (packer.PostProcessor, error) { return &shell_local_pp.PostProcessor{}, nil },
			"manifest":    func() (packer.PostProcessor, error) { return &manifest.PostProcessor{}, nil },
		},
	}
}
