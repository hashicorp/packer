package command

import (
	"flag"
)

// BuildOptionFlags sets the proper command line flags needed for
// build options.
func BuildOptionFlags(fs *flag.FlagSet, f *BuildOptions) {
	fs.Var((*SliceValue)(&f.Except), "except", "build all builds except these")
	fs.Var((*SliceValue)(&f.Only), "only", "only build the given builds by name")
}
