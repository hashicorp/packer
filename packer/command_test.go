package packer

type TestCommand struct {
	runArgs   []string
	runCalled bool
	runEnv    Environment
}

func (tc *TestCommand) Help() string {
	return "bar"
}

func (tc *TestCommand) Run(env Environment, args []string) int {
	tc.runCalled = true
	tc.runArgs = args
	tc.runEnv = env
	return 0
}

func (tc *TestCommand) Synopsis() string {
	return "foo"
}
