package command

import "context"

// PackerInterface is the interface to use packer; it represents ways users can
// use Packer. A call returns a int that will be the exit code of Packer,
// everything else is up to the implementer.
type PackerInterface interface {
	Build(ctx context.Context, args *BuildArgs) int
	Console(ctx context.Context, args *ConsoleArgs) int
	Fix(ctx context.Context, args *FixArgs) int
	Validate(ctx context.Context, args *ValidateArgs) int
}
