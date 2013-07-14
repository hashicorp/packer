package fix

import (
	"bytes"
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

	// Run the template through the various fixers
	fixers := []Fixer{Fixers["iso-md5"]}
	input := templateData
	for _, fixer := range fixers {
		var err error
		input, err = fixer.Fix(input)
		if err != nil {
			env.Ui().Error(fmt.Sprintf("Error fixing: %s", err))
			return 1
		}
	}

	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	if err := encoder.Encode(input); err != nil {
		env.Ui().Error(fmt.Sprintf("Error encoding: %s", err))
		return 1
	}

	var indented bytes.Buffer
	if err := json.Indent(&indented, output.Bytes(), "", "  "); err != nil {
		env.Ui().Error(fmt.Sprintf("Error encoding: %s", err))
		return 1
	}

	result := indented.String()
	result = strings.Replace(result, `\u003c`, "<", -1)
	result = strings.Replace(result, `\u003e`, ">", -1)
	env.Ui().Say(result)
	return 0
}

func (c Command) Synopsis() string {
	return "fixes templates from old versions of packer"
}
