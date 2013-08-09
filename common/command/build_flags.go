package command

import (
	"flag"
)

// BuildFilterFlags sets the proper command line flags needed for
// build filters.
func BuildFilterFlags(fs *flag.FlagSet, f *BuildFilters) {
	fs.Var((*SliceValue)(&f.Except), "except", "build all builds except these")
	fs.Var((*SliceValue)(&f.Only), "only", "only build the given builds by name")
}
