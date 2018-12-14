// Package sed implements the classic UNIX sed language in pure Go.
// The interface is very simple: a user compiles a program into an
// execution engine by calling New or NewQuiet. Then, the engine
// can Wrap() any io.Reader to lazily process the stream as you
// read from it.
//
// All classic sed commands are supported, but since the package
// uses Go's regexp package for the regular expressions, the syntax
// for regexps will not be the same as a typical UNIX sed.  In other
// words, instead of:  s|ab\(c*\)d|\1|g  you would say: s|ab(c*)d|$1|g.
// So this is a Go-flavored sed, rather than a drop-in replacement for
// a UNIX sed.  Depending on your tastes, you will either consider this
// an improvement or completely brain-dead.
package sed

import (
	"bufio"
	"bytes"
        "strings"
	"io"
)

// Engine is the compiled instruction stream for a sed program.
// It is the main type that users of the go-sed library will
// interact with.
type Engine struct {
	ins []instruction // the instruction stream
}

// vm is the virtual machine state for a running sed program.
type vm struct {
	nxtl     string        // the next line
	pat      string        // the pattern space, possibly nil
	hold     string        // the hold buffer,   possibly nil
	appl     *string       // any lines we've been asked to 'a\'ppend, usually nil
	overflow string        // any overflow we might have accumulated
	lastl    bool          // true if it's the last line
	ins      []instruction // the instruction stream
	ip       int           // the current locaiton in the instruction stream
	input    *bufio.Reader // the input stream
	output   []byte        // the output buffer
	lineno   int           // current line number
	modified bool          // have we modified the pattern space?
}

// a sed instruction is mostly a function transforming an engine
type instruction func(*vm) error

// makeEngine is the logic behine the New and NewQuiet public functions.
// It lexes and parses the program, and makes a new Engine out of it.
func makeEngine(program io.Reader, isQuiet bool) (*Engine, error) {
	bufprog := bufio.NewReader(program)
	ch := make(chan *token, 128)
	errch := make(chan error, 1)
	go lex(bufprog, ch, errch)

	instructions, parseErr := parse(ch, isQuiet)
	var err = <-errch // look for lexing errors first...
	if err == nil {
		// if there were no lex errors, look for a parsing error
		err = parseErr
	}

	return &Engine{ins: instructions}, err
}

// New creates a new sed engine from a program.  The program is executed
// via the Run method. If the provided program has any errors, the returned
// engine will be 'nil' and the error will be returned.  Otherwise, the returned
// error will be nil.
func New(program io.Reader) (*Engine, error) {
	return makeEngine(program, false)
}

// NewQuiet creates a new sed engine from a program.  It behaves exactly as
// New(), except it produces an engine that doesn't print lines by defualt. This
// is the classic '-n' sed behaviour.
func NewQuiet(program io.Reader) (*Engine, error) {
	return makeEngine(program, true)
}

// Wrap supplies an io.Reader that applies the sed Engine to the given
// input.  The sed program is run lazily against the input as the user
// asks for bytes.  If you'd prefer to run all at once from string to
// string, use RunString instead.
func (e *Engine) Wrap(input io.Reader) io.Reader {
	bufin := bufio.NewReader(input)

	// prime the engine by resetting the internal flags and filling nxtl...
	return &vm{ins: e.ins, input: bufin, lineno: -1, ip: -1}
}

// Read turns a vm into an io.Reader.
func (v *vm) Read(p []byte) (int, error) {
	var err error
	v.output = p

	if v.lineno == -1 {
		// we have an uninitialized stream
		err = cmd_fillNext(v)
		v.ip = 0
	} else if len(v.overflow) > 0 {
		// we have overflow to work on
		o := v.overflow
		v.overflow = ""
		err = writeString(v, o)
	}

	// run the program
	for err == nil {
		err = v.ins[v.ip](v)
	}

	var n int = len(p) - len(v.output)

	if ((err == fullBuffer) || (err == io.EOF)) && (n > 0) {
		err = nil
	}

	return n, err
}

// RunString executes the program embodied by the Engine on the
// given string as input, returning the output string and any
// errors that occured.
func (e *Engine) RunString(input string) (string, error) {
	inbuf := strings.NewReader(input)
	var outbytes bytes.Buffer

	_, err := io.Copy(&outbytes, e.Wrap(inbuf))

	if err == io.EOF {
		err = nil
	}

	return outbytes.String(), err
}
