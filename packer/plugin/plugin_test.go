package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func helperProcess(s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	env := []string{
		"GO_WANT_HELPER_PROCESS=1",
		"PACKER_PLUGIN_MIN_PORT=10000",
		"PACKER_PLUGIN_MAX_PORT=25000",
	}

	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(env, os.Environ()...)
	return cmd
}

// This is not a real test. This is just a helper process kicked off by
// tests.
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
	case "builder":
		ServeBuilder(new(helperBuilder))
	case "command":
		ServeCommand(new(helperCommand))
	case "hook":
		ServeHook(new(helperHook))
	case "invalid-rpc-address":
		fmt.Println("lolinvalid")
	case "mock":
		fmt.Println(":1234")
		<-make(chan int)
	case "post-processor":
		ServePostProcessor(new(helperPostProcessor))
	case "provisioner":
		ServeProvisioner(new(helperProvisioner))
	case "start-timeout":
		time.Sleep(1 * time.Minute)
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n", cmd)
		os.Exit(2)
	}
}
