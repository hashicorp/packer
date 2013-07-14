package fix

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"os"
	"strings"
)

type Command byte

func (Command) Help() string {
	return strings.TrimSpace(helpString)
}

func (c Command) Run(env packer.Environment, args []string) int {
	cmdFlags := flag.NewFlagSet("fix", flag.ContinueOnError)
	cmdFlags.Usage = func() { env.Ui().Say(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	args = cmdFlags.Args()
	if len(args) != 1 {
		cmdFlags.Usage()
		return 1
	}

	// Read the file for decoding
	tplF, err := os.Open(args[0])
	if err != nil {
		env.Ui().Error(fmt.Sprintf("Error opening template: %s", err))
		return 1
	}
	defer tplF.Close()

	// Decode the JSON into a generic map structure
	var templateData map[string]interface{}
	decoder := json.NewDecoder(tplF)
	if err := decoder.Decode(&templateData); err != nil {
		env.Ui().Error(fmt.Sprintf("Error parsing template: %s", err))
		return 1
	}

	// Close the file since we're done with that
	tplF.Close()

	return 0
}

func (c Command) Synopsis() string {
	return "fixes templates from old versions of packer"
}
