package check

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
)

type Stream int

const (
	// BothStreams will use both stdout and stderr for performing a check
	BothStreams Stream = iota
	// OnlyStdout will only use stdout for performing a check
	OnlyStdout
	// OnlySterr will only use stderr for performing a check
	OnlyStderr
)

func (s Stream) String() string {
	switch s {
	case BothStreams:
		return "Both streams"
	case OnlyStdout:
		return "Stdout"
	case OnlyStderr:
		return "Stderr"
	}

	panic(fmt.Sprintf("Unknown stream value: %d", s))
}

// Checker represents anything that can be used in conjunction with Assert.
//
// The role of a checker is performing a test on a command's outputs/error, and
// return an error if the test fails.
//
// Note: the Check method is the only required, however during tests the name
// of the checker is printed out in case it fails, so it may be useful to have
// a custom string for this: the `Name() string` method is exactly what to
// implement for this kind of customization.
type Checker interface {
	Check(stdout, stderr string, err error) error
}

func InferName(c Checker) string {
	if c == nil {
		panic("nil checker - malformed test?")
	}

	checkerType := reflect.TypeOf(c)
	_, ok := checkerType.MethodByName("Name")
	if !ok {
		return checkerType.String()
	}

	retVals := reflect.ValueOf(c).MethodByName("Name").Call([]reflect.Value{})
	if len(retVals) != 1 {
		panic(fmt.Sprintf("Name function called - returned %d values. Must be one string only.", len(retVals)))
	}

	return retVals[0].String()
}

func MustSucceed() Checker {
	return mustSucceed{}
}

type mustSucceed struct{}

func (_ mustSucceed) Check(stdout, stderr string, err error) error {
	return err
}

func MustFail() Checker {
	return mustFail{}
}

type mustFail struct{}

func (_ mustFail) Check(stdout, stderr string, err error) error {
	if err == nil {
		return fmt.Errorf("unexpected command success")
	}
	return nil
}

type GrepOpts int

const (
	// Invert the check, i.e. by default an empty grep fails, if this is set, a non-empty grep fails
	GrepInvert GrepOpts = iota
	// Only grep stderr
	GrepStderr
	// Only grep stdout
	GrepStdout
)

// Grep returns a checker that performs a regexp match on the command's output and returns an error if it failed
//
// Note: by default both streams will be checked by the grep
func Grep(expression string, opts ...GrepOpts) Checker {
	pc := PipeChecker{
		name:   fmt.Sprintf("command | grep -E %q", expression),
		stream: BothStreams,
		pipers: []Pipe{
			PipeGrep(expression),
		},
		check: ExpectNonEmptyInput(),
	}
	for _, opt := range opts {
		switch opt {
		case GrepInvert:
			pc.check = ExpectEmptyInput()
		case GrepStderr:
			pc.stream = OnlyStderr
		case GrepStdout:
			pc.stream = OnlyStdout
		}
	}
	return pc
}

func GrepInverted(expression string, opts ...GrepOpts) Checker {
	return Grep(expression, append(opts, GrepInvert)...)
}

type PluginVersionTuple struct {
	Source  string
	Version *version.Version
}

func NewPluginVersionTuple(src, pluginVersion string) PluginVersionTuple {
	ver := version.Must(version.NewVersion(pluginVersion))
	return PluginVersionTuple{
		Source:  src,
		Version: ver,
	}
}

type pluginsUsed struct {
	invert  bool
	plugins []PluginVersionTuple
}

func (pu pluginsUsed) Check(stdout, stderr string, _ error) error {
	var opts []GrepOpts
	if !pu.invert {
		opts = append(opts, GrepInvert)
	}

	var retErr error

	for _, pvt := range pu.plugins {
		// `error` is ignored for Grep, so we can pass in nil
		err := Grep(
			fmt.Sprintf("%s_v%s[^:]+\\\\s*plugin process exited", pvt.Source, pvt.Version.Core()),
			opts...,
		).Check(stdout, stderr, nil)
		if err != nil {
			retErr = multierror.Append(retErr, err)
		}
	}

	return retErr
}

// PluginsUsed is a glorified `Grep` checker that looks for a bunch of plugins
// used from the logs of a packer build or packer validate.
//
// Each tuple passed as parameter is looked for in the logs using Grep
func PluginsUsed(invert bool, plugins ...PluginVersionTuple) Checker {
	return pluginsUsed{
		invert:  invert,
		plugins: plugins,
	}
}

func Dump(t *testing.T) Checker {
	return &dump{t}
}

type dump struct {
	t *testing.T
}

func (d dump) Check(stdout, stderr string, err error) error {
	d.t.Logf("Dumping command result.")
	d.t.Logf("Stdout: %s", stdout)
	d.t.Logf("stderr: %s", stderr)
	return nil
}

type PanicCheck struct{}

func (_ PanicCheck) Check(stdout, stderr string, _ error) error {
	if strings.Contains(stdout, "= PACKER CRASH =") || strings.Contains(stderr, "= PACKER CRASH =") {
		return fmt.Errorf("packer has crashed: this is never normal and should be investigated")
	}
	return nil
}

// CustomCheck is meant to be a one-off checker with a user-provided function.
//
// Use this if none of the existing checkers match your use case, and it is not
// reusable/generic enough for use in other tests.
type CustomCheck struct {
	name      string
	checkFunc func(stdout, stderr string, err error) error
}

func (c CustomCheck) Check(stdout, stderr string, err error) error {
	return c.checkFunc(stdout, stderr, err)
}

func (c CustomCheck) Name() string {
	return fmt.Sprintf("custom check - %s", c.name)
}

// LineCountCheck builds a pipe checker to count the number of lines on stdout by default
//
// To change the stream(s) on which to perform the check, you can call SetStream on the
// returned pipe checker.
func LineCountCheck(lines int) *PipeChecker {
	return MkPipeCheck(fmt.Sprintf("line count (%d)", lines), LineCount()).
		SetTester(IntCompare(Eq, lines)).
		SetStream(OnlyStdout)
}
