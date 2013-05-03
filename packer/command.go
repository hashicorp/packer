package packer

// A command is a runnable sub-command of the `packer` application.
// When `packer` is called with the proper subcommand, this will be
// called.
//
// The mapping of command names to command interfaces is in the
// Environment struct.
//
// Run should run the actual command with the given environmet and
// command-line arguments. It should return the exit status when it is
// finished.
//
// Synopsis should return a one-line, short synopsis of the command.
// This should be less than 50 characters ideally.
type Command interface {
	Run(env Environment, args []string) int
	Synopsis() string
}
