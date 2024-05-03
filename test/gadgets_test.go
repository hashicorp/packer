package test

import (
	"fmt"
	"reflect"
	"regexp"
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

type MustSucceed struct{}

func (_ MustSucceed) Check(stdout, stderr string, err error) error {
	return err
}

type MustFail struct{}

func (_ MustFail) Check(stdout, stderr string, err error) error {
	if err == nil {
		return fmt.Errorf("unexpected command success")
	}
	return nil
}

// Grep is essentially the equivalent to a normal grep -E on the command line.
//
// The `expect` string is meant to be a regexp, which will be compiled on-demand,
// and will panic if it isn't a valid POSIX extended regexp.
type Grep struct {
	streams Stream
	expect  string
}

func (g Grep) Check(stdout, stderr string, err error) error {
	re := regexp.MustCompilePOSIX(g.expect)

	streams := []string{}

	switch g.streams {
	case BothStreams:
		streams = append(streams, stdout, stderr)
	case OnlyStdout:
		streams = append(streams, stdout)
	case OnlyStderr:
		streams = append(streams, stderr)
	}

	var found bool
	for _, stream := range streams {
		found = found || re.MatchString(stream)
	}

	if !found {
		return fmt.Errorf("streams %q did not match the expected regexp %q", g.streams, g.expect)
	}
	return nil
}

func (g Grep) Name() string {
	return fmt.Sprintf("command (%s) | grep -E %q", g.streams, g.expect)
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
