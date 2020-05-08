package command

import (
	"flag"
	"fmt"
	"math"
	"strings"

	"github.com/hashicorp/packer/helper/enumflag"
	kvflag "github.com/hashicorp/packer/helper/flag-kv"
	sliceflag "github.com/hashicorp/packer/helper/flag-slice"
	"github.com/hashicorp/packer/packer"
)

// NewMetaArgs parses cli args and put possible values
func (ma *MetaArgs) AddFlagSets(fs *flag.FlagSet) {
	fs.Var((*sliceflag.StringFlag)(&ma.Only), "only", "")
	fs.Var((*sliceflag.StringFlag)(&ma.Except), "except", "")
	fs.Var((*kvflag.Flag)(&ma.Vars), "var", "")
	fs.Var((*kvflag.StringSlice)(&ma.VarFiles), "var-file", "")
}

// ConfigType tells what type of config we should use, it can return values
// like "hcl" or "json".
// Make sure Args was correctly set before.
func (ma *MetaArgs) ConfigType() (string, error) {
	switch len(ma.Args) {
	// TODO(azr): in the future, I want to allow passing multiple arguments to
	// merge HCL confs together; but this will probably need an RFC first.
	// TODO(azr): To allow piping HCL2 confs (when args is "-"), we probably
	// will need to add a setting that says "this is an HCL config".
	case 1:
		name := ma.Args[0]
		if strings.HasSuffix(name, ".pkr.hcl") ||
			strings.HasSuffix(name, ".pkr.json") {
			return "hcl", nil
		}
		isDir, err := isDir(name)
		if isDir {
			return "hcl", err
		}
		return "json", err
	default:
		return "", fmt.Errorf("packer only takes on argument: %q", ma.Args)
	}
}

// MetaArgs defines commonalities between all comands
type MetaArgs struct {
	Args         []string
	Only, Except []string
	Vars         map[string]string
	VarFiles     []string
	Ui           packer.Ui
}

func (ba *BuildArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.BoolVar(&ba.Color, "color", true, "")
	flags.BoolVar(&ba.Debug, "debug", false, "")
	flags.BoolVar(&ba.Force, "force", false, "")
	flags.BoolVar(&ba.TimestampUi, "timestamp-ui", false, "")
	flags.BoolVar(&ba.MachineReadable, "machine-readable", false, "")

	flags.Int64Var(&ba.ParallelBuilds, "parallel-builds", 0, "")

	flagOnError := enumflag.New(&ba.OnError, "cleanup", "abort", "ask")
	flags.Var(flagOnError, "on-error", "")

	ba.MetaArgs.AddFlagSets(flags)
}

func (ba *BuildArgs) ParseArgvs(args []string) int {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.Usage = func() { ba.Ui.Say(ba.Help()) }
	ba.AddFlagSets(flags)
	err := flags.Parse(args)
	if err != nil {
		return 1
	}

	if ba.ParallelBuilds < 1 {
		ba.ParallelBuilds = math.MaxInt64
	}

	ba.Args = flags.Args()
	return 0
}

// BuildArgs represents a parsed cli line for a `packer build`
type BuildArgs struct {
	MetaArgs
	Color, Debug, Force, TimestampUi, MachineReadable bool
	ParallelBuilds                                    int64
	OnError                                           string
}

func (ca *ConsoleArgs) ParseArgvs(args []string) int {
	flags := flag.NewFlagSet("console", flag.ContinueOnError)
	flags.Usage = func() { ca.Ui.Say(ca.Help()) }
	ca.AddFlagSets(flags)
	err := flags.Parse(args)
	if err != nil {
		return 1
	}

	ca.Args = flags.Args()
	return 0
}

// ConsoleArgs represents a parsed cli line for a `packer console`
type ConsoleArgs struct{ MetaArgs }

func (fa *FixArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.BoolVar(&fa.Validate, "validate", true, "")

	fa.MetaArgs.AddFlagSets(flags)
}

func (fa *FixArgs) ParseArgvs(args []string) int {
	flags := flag.NewFlagSet("fix", flag.ContinueOnError)
	flags.Usage = func() { fa.Ui.Say(fa.Help()) }
	fa.AddFlagSets(flags)
	err := flags.Parse(args)
	if err != nil {
		return 1
	}

	fa.Args = flags.Args()
	return 0
}

// FixArgs represents a parsed cli line for a `packer fix`
type FixArgs struct {
	MetaArgs
	Validate bool
}

func (va *ValidateArgs) AddFlagSets(flags *flag.FlagSet) {
	flags.BoolVar(&va.SyntaxOnly, "syntax-only", true, "")

	va.MetaArgs.AddFlagSets(flags)
}

func (va *ValidateArgs) ParseArgvs(args []string) int {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	flags.Usage = func() { va.Ui.Say(va.Help()) }
	va.AddFlagSets(flags)
	err := flags.Parse(args)
	if err != nil {
		return 1
	}

	va.Args = flags.Args()
	return 0
}

// ValidateArgs represents a parsed cli line for a `packer validate`
type ValidateArgs struct {
	MetaArgs
	SyntaxOnly bool
}
