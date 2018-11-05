package sed

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

var fullBuffer = errors.New("FullBuffer")

func writeString(svm *vm, str string) error {
	var err error
	end := len(svm.output)
	src := str
	srclen := len(src)
	if end < srclen {
		src = src[:end]
		srclen = end
		svm.overflow += str[end:]
		err = fullBuffer
	}
	for i := 0; i < srclen; i++ {
		svm.output[i] = src[i]
	}

	svm.output = svm.output[srclen:]
	return err
}

func cmd_quit(svm *vm) error {
	return io.EOF
}

// ---------------------------------------------------
func cmd_swap(svm *vm) error {
	svm.pat, svm.hold = svm.hold, svm.pat
	svm.ip++
	return nil
}

// ---------------------------------------------------
func cmd_get(svm *vm) error {
	svm.pat = svm.hold
	svm.ip++
	return nil
}

// ---------------------------------------------------
func cmd_hold(svm *vm) error {
	svm.hold = svm.pat
	svm.ip++
	return nil
}

// ---------------------------------------------------
func cmd_getapp(svm *vm) error {
	svm.pat = strings.Join([]string{svm.pat, svm.hold}, "\n")
	svm.ip++
	return nil
}

// ---------------------------------------------------
func cmd_holdapp(svm *vm) error {
	svm.hold = strings.Join([]string{svm.hold, svm.pat}, "\n")
	svm.ip++
	return nil
}

// ---------------------------------------------------
// newBranch generates branch instructions with specific
// targets
func cmd_newBranch(target int) instruction {
	return func(svm *vm) error {
		svm.ip = target
		return nil
	}
}

// ---------------------------------------------------
// newChangedBranch generates branch instructions with specific
// targets that only trigger on modified pattern spaces
func cmd_newChangedBranch(target int) instruction {
	return func(svm *vm) error {
		if svm.modified {
			svm.ip = target
			svm.modified = false
		} else {
			svm.ip++
		}
		return nil
	}
}

// ---------------------------------------------------
func cmd_print(svm *vm) error {
	svm.ip++

	writeString(svm, svm.pat)
	return writeString(svm, "\n")
}

// ---------------------------------------------------
func cmd_printFirstLine(svm *vm) error {
	svm.ip++

	idx := strings.IndexRune(svm.pat, '\n')

	if idx == -1 {
		idx = len(svm.pat)
	}

	writeString(svm, svm.pat[:idx])
	return writeString(svm, "\n")
}

// ---------------------------------------------------
func cmd_deleteFirstLine(svm *vm) (err error) {
	idx := strings.IndexRune(svm.pat, '\n')

	if idx == -1 {
		svm.pat = ""
		svm.ip = 0 // go back and fillNext
	} else {
		svm.pat = svm.pat[idx+1:]
		svm.ip = 1 // restart, but skip filling
	}

	return nil
}

// ---------------------------------------------------
func cmd_lineno(svm *vm) error {
	svm.ip++
	var lineno = fmt.Sprintf("%d\n", svm.lineno)
	return writeString(svm, lineno)
}

// ---------------------------------------------------
func cmd_fillNext(svm *vm) error {
	var err error

	// first, put out any stored-up 'a\'ppended text:
	if svm.appl != nil {
		err = writeString(svm, *svm.appl)
		svm.appl = nil
		if err != nil {
			return err // ok, since IP unchanged
		}
	}

	// just return if we're at EOF
	if svm.lastl {
		return io.EOF
	}

	// otherwise, copy nxtl to the pattern space and
	// refill.
	svm.ip++

	svm.pat = svm.nxtl
	svm.lineno++
	svm.modified = false

	var prefix = true
	var line []byte

	var lines []string

	for prefix {
		line, prefix, err = svm.input.ReadLine()
		if err != nil {
			break
		}
		// buf := make([]byte, len(line))
		// copy(buf, line)
		lines = append(lines, string(line))
	}

	svm.nxtl = strings.Join(lines, "")

	if err == io.EOF {
		if len(svm.nxtl) == 0 {
			svm.lastl = true
		}
		err = nil
	}

	return err
}

func cmd_fillNextAppend(svm *vm) error {
	var lines = make([]string, 2)
	lines[0] = svm.pat
	err := cmd_fillNext(svm) // increments svm.ip, so we don't
	lines[1] = svm.pat
	svm.pat = strings.Join(lines, "\n")
	return err
}

// --------------------------------------------------

type cmd_simplecond struct {
	cond     condition // the condition to check
	metloc   int       // where to jump if the condition is met
	unmetloc int       // where to jump if the condition is not met
}

func (c *cmd_simplecond) run(svm *vm) error {
	if c.cond.isMet(svm) {
		svm.ip = c.metloc
	} else {
		svm.ip = c.unmetloc
	}
	return nil
}

// --------------------------------------------------
type cmd_twocond struct {
	start    condition // the condition that begines the block
	end      condition // the condition that ends the block
	metloc   int       // where to jump if the condition is met
	unmetloc int       // where to jump if the condition is not met
	isOn     bool      // are we active already?
	offFrom  int       // if we saw the end condition, what line was it on?
}

func newTwoCond(c1 condition, c2 condition, metloc int, unmetloc int) *cmd_twocond {
	return &cmd_twocond{c1, c2, metloc, unmetloc, false, 0}
}

// isLastLine is here to support multi-line "c\" commands.
// The command needs to know when it's the end of the
// section so it can do the replacement.
func (c *cmd_twocond) isLastLine(svm *vm) bool {
	return c.isOn && (c.offFrom == svm.lineno)
}

func (c *cmd_twocond) run(svm *vm) error {
	if c.isOn && (c.offFrom > 0) && (c.offFrom < svm.lineno) {
		c.isOn = false
		c.offFrom = 0
	}

	if !c.isOn {
		if c.start.isMet(svm) {
			svm.ip = c.metloc
			c.isOn = true
		} else {
			svm.ip = c.unmetloc
		}
	} else {
		if c.end.isMet(svm) {
			c.offFrom = svm.lineno
		}
		svm.ip = c.metloc
	}
	return nil
}

// --------------------------------------------------
func cmd_newChanger(text string, guard *cmd_twocond) instruction {
	return func(svm *vm) error {
		svm.ip = 0 // go to the the next cycle

		var err error
		if (guard == nil) || guard.isLastLine(svm) {
			err = writeString(svm, text)
		}
		return err
	}
}

// --------------------------------------------------
func cmd_newAppender(text string) instruction {
	return func(svm *vm) error {
		svm.ip++
		if svm.appl == nil {
			svm.appl = &text
		} else {
			var newstr = *svm.appl + text
			svm.appl = &newstr
		}
		return nil
	}
}

// --------------------------------------------------
func cmd_newInserter(text string) instruction {
	return func(svm *vm) error {
		svm.ip++
		return writeString(svm, text)
	}
}

// --------------------------------------------------
// The 'r' command is basically and 'a\' with the contents
// of a filsvm. I implement it literally that way below.
func cmd_newReader(filename string) (instruction, error) {
	bytes, err := ioutil.ReadFile(filename)
	return cmd_newAppender(string(bytes)), err
}

// --------------------------------------------------
// The 'w' command appends the current pattern space
// to the named filsvm.  In this implementation, it opens
// the file for appending, writes the file, and then
// closes the filsvm.  This appears to be consistent with
// what OS X sed does.
func cmd_newWriter(filename string) instruction {
	return func(svm *vm) error {
		svm.ip++
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			defer f.Close()
			_, err = f.WriteString(svm.pat)
		}
		if err == nil {
			_, err = f.WriteString("\n")
		}
		return err
	}
}
