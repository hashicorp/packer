package test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
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

type grepOpts int

const (
	// Invert the check, i.e. by default an empty grep fails, if this is set, a non-empty grep fails
	grepInvert grepOpts = iota
	// Only grep stderr
	grepStderr
	// Only grep stdout
	grepStdout
)

// Grep returns a checker that performs a regexp match on the command's output and returns an error if it failed
//
// Note: by default both streams will be checked by the grep
func Grep(expression string, opts ...grepOpts) Checker {
	pc := PipeChecker{
		name:   "command | grep -E %q",
		stream: BothStreams,
		pipers: []Pipe{
			PipeGrep(expression),
		},
		check: ExpectNonEmptyInput(),
	}
	for _, opt := range opts {
		switch opt {
		case grepInvert:
			pc.check = ExpectEmptyInput()
		case grepStderr:
			pc.stream = OnlyStderr
		case grepStdout:
			pc.stream = OnlyStdout
		}
	}
	return pc
}

type Dump struct {
	t *testing.T
}

func (d Dump) Check(stdout, stderr string, err error) error {
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
