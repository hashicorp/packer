package test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Pipe is any command that allows piping two gadgets together
//
// There's always only one input and one output (stdout), mimicking essentially
// how pipes work in the UNIX-world.
type Pipe interface {
	Process(input string) (string, error)
}

// CustomPipe allows providing a simple function for piping inputs together
type CustomPipe func(string) (string, error)

func (c CustomPipe) Process(input string) (string, error) {
	return c(input)
}

// PipeGrep performs a grep on an input and returns the matches, one-per-line.
//
// The expression passed as parameter will be compiled to a POSIX extended regexp.
func PipeGrep(expression string) Pipe {
	re := regexp.MustCompilePOSIX(expression)
	return CustomPipe(func(input string) (string, error) {
		return strings.Join(re.FindAllString(input, -1), "\n"), nil
	})
}

// LineCount counts the number of lines received
//
// This excludes empty lines.
func LineCount() Pipe {
	return CustomPipe(func(s string) (string, error) {
		lines := strings.FieldsFunc(s, func(r rune) bool {
			return r == '\n'
		})
		return fmt.Sprintf("%d\n", len(lines)), nil
	})
}

// Tester is the end of a pipe for testing purposes.
//
// Once multiple commands have been piped together in a pipeline, we can
// perform some checks on that input, and decide if a test is a success or a
// failure.
type Tester interface {
	Check(input string) error
}

// CustomTester allows providing a function to check that the input is what we want
type CustomTester func(string) error

func (ct CustomTester) Check(input string) error {
	return ct(input)
}

// ExpectNonEmptyInput errors if the result from the pipeline was empty
//
// Non-empty in this context means that the output contains characters that are
// non-whitespace, i.e. anything that `TrimSpace` (aka unicode.IsSpace) recognizes
// as whitespace.
func ExpectNonEmptyInput() Tester {
	return CustomTester(func(in string) error {
		in = strings.TrimSpace(in)
		if in == "" {
			return fmt.Errorf("input is empty, expected it not to")
		}
		return nil
	})
}

// ExpectEmptyInput errors if the result from the pipeline was not empty
//
// Non-empty in this context means that the output contains characters that are
// non-whitespace, i.e. anything that `TrimSpace` (aka unicode.IsSpace) recognizes
// as whitespace.
func ExpectEmptyInput() Tester {
	return CustomTester(func(in string) error {
		in = strings.TrimSpace(in)
		if in != "" {
			return fmt.Errorf("input is not empty, expected it to be: %s", in)
		}
		return nil
	})
}

type op int

const (
	eq op = iota
	ne
	gt
	ge
	lt
	le
)

func (op op) String() string {
	switch op {
	case eq:
		return "=="
	case ne:
		return "!="
	case gt:
		return ">"
	case ge:
		return ">="
	case lt:
		return "<"
	case le:
		return "<="
	}

	panic(fmt.Sprintf("Unknown operator %d", op))
}

// IntCompare reads the input from the pipeline and compares it to a value using `op`
//
// If the input is not an int, this fails.
func IntCompare(op op, value int) Tester {
	return CustomTester(func(in string) error {
		n, err := strconv.Atoi(in)
		if err != nil {
			return fmt.Errorf("not an integer %q: %s", in, err)
		}

		var result bool
		switch op {
		case eq:
			result = n == value
		case ne:
			result = n != value
		case gt:
			result = n > value
		case ge:
			result = n >= value
		case lt:
			result = n < value
		case le:
			result = n <= value
		default:
			panic(fmt.Sprintf("Unsupported operator %d, make sure the operation is implemented for IntCompare", op))
		}

		if !result {
			return fmt.Errorf("comparison failed: %d %s %d -> %t", n, op, value, result)
		}

		return nil
	})
}

// PipeChecker is a kind of checker that essentially lets users write mini
// gadgets that pipe inputs together, and compose those to end as a true/false
// statement, which translates to an error.
//
// Unlike pipes in a real command-line context, since we're dealing with
// finite gadgets to process data, we're sequentially running their Process
// function, and any processor that ends in an error will make the pipeline
// fail.
//
// Stream is provided so we know if we want to combine stdout/stderr for the
// pipeline, or if we want only to focus on either.
type PipeChecker struct {
	name   string
	stream Stream

	pipers []Pipe
	check  Tester
}

func (pc PipeChecker) Check(stdout, stderr string, _ error) error {
	if len(pc.pipers) == 0 {
		return fmt.Errorf("%s - empty pipeline", pc.Name())
	}

	var pipeStr string
	switch pc.stream {
	case OnlyStdout:
		pipeStr = stdout
	case OnlyStderr:
		pipeStr = stderr
	case BothStreams:
		pipeStr = fmt.Sprintf("%s\n%s", stdout, stderr)
	}

	var err error
	for i, pp := range pc.pipers {
		pipeStr, err = pp.Process(pipeStr)
		if err != nil {
			return fmt.Errorf("pipeline failed during execution (%d): %s", i, err)
		}
	}
	return pc.check.Check(pipeStr)
}

func (pc PipeChecker) Name() string {
	rawName := "|>?"
	if pc.name != "" {
		rawName = fmt.Sprintf("%s - %s", rawName, pc.name)
	}
	return rawName
}
