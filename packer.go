// This is the main package for the `packer` application.
package main

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"runtime"
)

func loadGlobalConfig() (result *config, err error) {
	mustExist := true
	p := os.Getenv("PACKER_CONFIG")
	if p == "" {
		var u *user.User
		u, err = user.Current()
		if err != nil {
			return
		}

		p = path.Join(u.HomeDir, ".packerrc")
		mustExist = false
	}

	log.Printf("Loading packer config: %s\n", p)
	contents, err := ioutil.ReadFile(p)
	if err != nil && !mustExist {
		// Don't report an error if it is okay if the file is missing
		perr, ok := err.(*os.PathError)
		if ok && perr.Op == "open" {
			log.Printf("Packer config didn't exist. Ignoring: %s\n", p)
			err = nil
		}
	}

	if err != nil {
		return
	}

	result, err = parseConfig(string(contents))
	if err != nil {
		return
	}

	return
}

func main() {
	if os.Getenv("PACKER_LOG") == "" {
		// If we don't have logging explicitly enabled, then disable it
		log.SetOutput(ioutil.Discard)
	} else {
		// Logging is enabled, make sure it goes to stderr
		log.SetOutput(os.Stderr)
	}

	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	defer plugin.CleanupClients()

	homeConfig, err := loadGlobalConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading global Packer configuration: \n\n%s\n", err)
		os.Exit(1)
	}

	config, err := parseConfig(defaultConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing global Packer configuration: \n\n%s\n", err)
		os.Exit(1)
	}

	if homeConfig != nil {
		log.Println("Merging default config with home config...")
		config = mergeConfig(config, homeConfig)
	}

	envConfig := packer.DefaultEnvironmentConfig()
	envConfig.Commands = config.CommandNames()
	envConfig.Components.Builder = config.LoadBuilder
	envConfig.Components.Command = config.LoadCommand

	env, err := packer.NewEnvironment(envConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Packer initialization error: \n\n%s\n", err)
		os.Exit(1)
	}

	exitCode, err := env.Cli(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}

	plugin.CleanupClients()
	os.Exit(exitCode)
}
