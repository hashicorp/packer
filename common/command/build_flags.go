package command

import (
	"flag"
	"fmt"
	"strings"
)

// BuildOptionFlags sets the proper command line flags needed for
// build options.
func BuildOptionFlags(fs *flag.FlagSet, f *BuildOptions) {
	fs.Var((*SliceValue)(&f.Except), "except", "build all builds except these")
	fs.Var((*SliceValue)(&f.Only), "only", "only build the given builds by name")
	fs.Var((*userVarValue)(&f.UserVars), "var", "specify a user variable")
	fs.Var((*AppendSliceValue)(&f.UserVarFiles), "var-file", "file with user variables")
}

// userVarValue is a flag.Value that parses out user variables in
// the form of 'key=value' and sets it on this map.
type userVarValue map[string]string

func (v *userVarValue) String() string {
	return ""
}

func (v *userVarValue) Set(raw string) error {
	idx := strings.Index(raw, "=")
	if idx == -1 {
		return fmt.Errorf("No '=' value in arg: %s", raw)
	}

	if *v == nil {
		*v = make(map[string]string)
	}

	key, value := raw[0:idx], raw[idx+1:]
	(*v)[key] = value
	return nil
}
