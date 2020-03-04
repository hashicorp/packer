package command

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/hashicorp/packer/helper/wrappedreadline"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/posener/complete"
)

const TiniestBuilder = `{
	"builders": [
		{
			"type":"null",
			"communicator": "none"
		}
	]
}`

type ConsoleCommand struct {
	Meta
}

func (c *ConsoleCommand) Run(args []string) int {
	flags := c.Meta.FlagSet("console", FlagSetVars)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return 1
	}

	var templ *template.Template

	args = flags.Args()
	if len(args) < 1 {
		// If user has not defined a builder, create a tiny null placeholder
		// builder so that we can properly initialize the core
		tpl, err := template.Parse(strings.NewReader(TiniestBuilder))
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to generate placeholder template: %s", err))
			return 1
		}
		templ = tpl
	} else if len(args) == 1 {
		// Parse the provided template
		tpl, err := template.ParseFile(args[0])
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to parse template: %s", err))
			return 1
		}
		templ = tpl
	} else {
		// User provided too many arguments
		flags.Usage()
		return 1
	}

	// Get the core
	core, err := c.Meta.Core(templ)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	// IO Loop
	session := &REPLSession{
		Core: core,
	}

	// Determine if stdin is a pipe. If so, we evaluate directly.
	if c.StdinPiped() {
		return c.modePiped(session)
	}

	return c.modeInteractive(session)
}

func (*ConsoleCommand) Help() string {
	helpText := `
Usage: packer console [options] [TEMPLATE]

  Creates a console for testing variable interpolation.
  If a template is provided, this command will load the template and any
  variables defined therein into its context to be referenced during
  interpolation.

Options:
  -var 'key=value'       Variable for templates, can be used multiple times.
  -var-file=path         JSON file containing user variables. [ Note that even in HCL mode this expects file to contain JSON, a fix is comming soon ]
`

	return strings.TrimSpace(helpText)
}

func (*ConsoleCommand) Synopsis() string {
	return "creates a console for testing variable interpolation"
}

func (*ConsoleCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*ConsoleCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-var":      complete.PredictNothing,
		"-var-file": complete.PredictNothing,
	}
}

func (c *ConsoleCommand) modePiped(session *REPLSession) int {
	var lastResult string
	scanner := bufio.NewScanner(wrappedreadline.Stdin())
	for scanner.Scan() {
		result, err := session.Handle(strings.TrimSpace(scanner.Text()))
		if err != nil {
			return 0
		}
		// Store the last result
		lastResult = result
	}

	// Output the final result
	c.Ui.Message(lastResult)
	return 0
}

func (c *ConsoleCommand) modeInteractive(session *REPLSession) int { // Setup the UI so we can output directly to stdout
	l, err := readline.NewEx(wrappedreadline.Override(&readline.Config{
		Prompt:            "> ",
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	}))
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Error initializing console: %s",
			err))
		return 1
	}
	for {
		// Read a line
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		out, err := session.Handle(line)
		if err == ErrSessionExit {
			break
		}
		if err != nil {
			c.Ui.Error(err.Error())
			continue
		}

		c.Ui.Say(out)
	}

	return 0
}

// ErrSessionExit is a special error result that should be checked for
// from Handle to signal a graceful exit.
var ErrSessionExit = errors.New("Session exit")

// Session represents the state for a single Read-Evaluate-Print-Loop (REPL) session.
type REPLSession struct {
	// Core is used for constructing interpolations based off packer templates
	Core *packer.Core
}

// Handle a single line of input from the REPL.
//
// The return value is the output and the error to show.
func (s *REPLSession) Handle(line string) (string, error) {
	switch {
	case strings.TrimSpace(line) == "":
		return "", nil
	case strings.TrimSpace(line) == "exit":
		return "", ErrSessionExit
	case strings.TrimSpace(line) == "help":
		return s.handleHelp()
	case strings.TrimSpace(line) == "variables":
		return s.handleVariables()
	default:
		return s.handleEval(line)
	}
}

func (s *REPLSession) handleEval(line string) (string, error) {
	ctx := s.Core.Context()
	rendered, err := interpolate.Render(line, ctx)
	if err != nil {
		return "", fmt.Errorf("Error interpolating: %s", err)
	}
	return rendered, nil
}

func (s *REPLSession) handleVariables() (string, error) {
	varsstring := "\n"
	for k, v := range s.Core.Context().UserVariables {
		varsstring += fmt.Sprintf("%s: %+v,\n", k, v)
	}

	return varsstring, nil
}

func (s *REPLSession) handleHelp() (string, error) {
	text := `
The Packer console allows you to experiment with Packer interpolations.
You may access variables in the Packer config you called the console with.

Type in the interpolation to test and hit <enter> to see the result.

To exit the console, type "exit" and hit <enter>, or use Control-C.
`

	return strings.TrimSpace(text), nil
}
