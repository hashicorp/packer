package sed

// the lexer for SED.  The point of the lexer is to
// reliably transform the input into a series of token structs.
// These structs know the source location, and the token type, and
// any arguments to the token (e.g., a regexp's '/' argument is the
// regular expression itself).
//
// The lexer also simplifies and regularises the input, for instance
// by eliminating comments.

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type location struct {
	line int
	pos  int
}

func (l *location) String() string {
	return fmt.Sprintf("at line %d, pos %d", l.line, l.pos)
}

const (
	tok_NUM = iota
	tok_RX
	tok_COMMA
	tok_BANG
	tok_DOLLAR
	tok_LBRACE
	tok_RBRACE
	tok_EOL
	tok_CMD
	tok_CHANGE
	tok_LABEL
)

type token struct {
	location
	typ    int
	letter rune
	args   []string
}

// ----------------------------------------------------------
//  Location-tracking reader
// ----------------------------------------------------------
type locReader struct {
	location
	eol bool // state for end of line, true when last rune was '\n'
	r   *bufio.Reader
}

func (lr *locReader) ReadRune() (rune, int, error) {
	r, i, err := lr.r.ReadRune()

	lr.pos++

	if lr.eol {
		lr.pos = 1
		lr.line++
		lr.eol = false
	}
	if r == '\n' {
		lr.eol = true
	}

	return r, i, err
}

func (lr *locReader) UnreadRune() error {
	lr.pos--
	lr.eol = false

	if lr.pos == 0 {
		lr.line--
		lr.eol = true
	}
	return lr.r.UnreadRune()
}

func (lr *locReader) ReadLine() (nxtl string, err error) {
	var prefix = true
	var line []byte

	var lines []string

	for prefix {
		line, prefix, err = lr.r.ReadLine()
		if err != nil {
			break
		}
		buf := make([]byte, len(line))
		copy(buf, line)
		lines = append(lines, string(buf))
	}

	nxtl = strings.Join(lines, "")

	// fixup our position information
	lr.pos += len(nxtl)
	lr.eol = true

	return
}

// ----------------------------------------------------------
// lexer functions
// ----------------------------------------------------------
func skipComment(r *locReader) (rune, error) {
	var err error
	var cur rune = ' '
	for (cur != '\n') && (err == nil) {
		cur, _, err = r.ReadRune()
	}
	return ';', err
}

func skipWS(r *locReader) (rune, error) {
	var err error
	var cur rune = ' '
	for {
		switch {
		case cur == '\n':
			return ';', err
		case cur == '#':
			return skipComment(r)
		case !unicode.IsSpace(cur):
			return cur, err
		}
		cur, _, err = r.ReadRune()
	}
}

func readNumber(r *locReader, character rune) (string, error) {
	var buffer bytes.Buffer

	var err error
	for (err == nil) && unicode.IsDigit(character) {
		buffer.WriteRune(character)
		character, _, err = r.ReadRune()
	}

	if err == nil {
		err = r.UnreadRune()
	}

	return buffer.String(), err
}

// readDelimited reads until it finds the delimter character,
// returning the string (not including the delimiter). It does
// allow the delimiter to be escaped by a backslash ('\').
// It is an error to reach EOL while looking for the delimiter.
func readDelimited(r *locReader, delimiter rune) (string, error) {
	var buffer bytes.Buffer

	var err error
	var character rune
	var previous rune

	character, _, err = r.ReadRune()
	for (err == nil) &&
		(character != '\n') &&
		((character != delimiter) || (previous == '\\')) {
		buffer.WriteRune(character)
		previous = character
		character, _, err = r.ReadRune()
	}

	if character == '\n' {
		err = fmt.Errorf("end-of-line while looking for %c", delimiter)
	}

	if err == io.EOF {
		err = fmt.Errorf("end-of-file while looking for %c", delimiter)
	}

	return buffer.String(), err
}

// readReplacement reads until it finds the delimter character,
// returning the string (not including the delimiter). It does
// allow the delimiter to be escaped by a backslash ('\'), and it
// does interpret a few common backslash escapes like \n and \t.
// It is an error to reach an unescaped EOL while looking for the delimiter.
func readReplacement(r *locReader, delimiter rune) (string, error) {
	var buffer bytes.Buffer

	var err error
	var character rune
	var previous rune

	character, _, err = r.ReadRune()
	for err == nil {
		if character == '\r' {
			character, _, err = r.ReadRune()
			continue
		}

		if previous == '\\' {
			// find out what we escaped...
			switch character {
			case 'r':
				buffer.WriteRune('\r')
			case 't':
				buffer.WriteRune('\t')
			case 'n':
				buffer.WriteRune('\n')
			case '\\':
				buffer.WriteRune(character)
				character = ' ' // don't escape the next one
			default:
				buffer.WriteRune(character)
			}
		} else {
			if character == delimiter ||
				character == '\n' {
				break
			} else if character != '\\' {
				buffer.WriteRune(character)
			}
		}
		previous = character
		character, _, err = r.ReadRune()
	}

	if character == '\n' {
		err = fmt.Errorf("end-of-line while looking for %c", delimiter)
	}

	if err == io.EOF {
		err = fmt.Errorf("end-of-file while looking for %c", delimiter)
	}

	return buffer.String(), err
}

// readMultiLine reads until it finds an unescaped newline. It discards the
// first line, if it is empty, because commands like "c\", "a\" and "i\" are
// intended to be used that way.
func readMultiLine(r *locReader) (string, error) {
	var lines []string
	var err error

	first := true
	hasSlash := true // does the line end in a slash?

	for hasSlash {
		txt, err := r.ReadLine()
		if err != nil {
			break
		}
		tlen := len(txt)

		// strip off the final '\', if there is one
		if tlen > 0 && txt[tlen-1] == '\\' {
			txt = txt[:tlen-1]
		} else {
			hasSlash = false
		}

		// If it's empty and the first line, forget it.
		// Otherwise, add it to the line list
		if !first || tlen > 1 {
			lines = append(lines, txt)
		}

		first = false
	}

	// for sed's purposes, we want a final newline...
	lines = append(lines, "")

	return strings.Join(lines, "\n"), err
}

// readIdentifier skips any whitespace, and then reads until it
// finds either a ';' or a non-alphanumeric character.  It
// returns the string it reads.
func readIdentifier(r *locReader) (string, error) {
	var buffer bytes.Buffer

	var err error
	var character rune

	character, err = skipWS(r)
	for (err == nil) && (character != ';') && !unicode.IsSpace(character) {
		buffer.WriteRune(character)
		character, _, err = r.ReadRune()
	}

	if err == nil {
		err = r.UnreadRune()
	}
	return buffer.String(), err
}

func readSubstitution(r *locReader) ([]string, error) {
	var ans = make([]string, 3)
	var err error

	// step 1.: get the delimiter character for substitutions
	var delimiter rune
	delimiter, _, err = r.ReadRune()
	if err != nil {
		return ans, err
	}

	// step 2.: read the regexp
	ans[0], err = readDelimited(r, delimiter)
	if err != nil {
		return ans, err
	}

	// step 3.: read the replacement
	ans[1], err = readReplacement(r, delimiter)
	if err != nil {
		return ans, err
	}

	// step 4.: read the modifiers
	ans[2], err = readIdentifier(r)

	return ans, err
}

func readTranslation(r *locReader) ([]string, error) {
	var ans = make([]string, 2)
	var err error

	// step 1.: get the delimiter character for substitutions
	var delimiter rune
	delimiter, _, err = r.ReadRune()
	if err != nil {
		return ans, err
	}

	// step 2.: read the regexp
	ans[0], err = readDelimited(r, delimiter)
	if err != nil {
		return ans, err
	}

	// step 3.: read the replacement
	ans[1], err = readDelimited(r, delimiter)
	if err != nil {
		return ans, err
	}

	return ans, err
}

func lex(r *bufio.Reader, ch chan<- *token, errch chan<- error) {
	defer close(ch)
	defer close(errch)

	rdr := locReader{}
	rdr.r = r
	rdr.eol = true

	var err error
	var cur rune

	var topLoc = rdr.location

	for err == nil {
		cur, err = skipWS(&rdr)
		if err != nil {
			break
		}

		topLoc = rdr.location // remember the start of the command

		switch cur {
		case ';':
			ch <- &token{topLoc, tok_EOL, cur, nil}
		case ',':
			ch <- &token{topLoc, tok_COMMA, cur, nil}
		case '{':
			ch <- &token{topLoc, tok_LBRACE, cur, nil}
		case '}':
			ch <- &token{topLoc, tok_RBRACE, cur, nil}
		case '!':
			ch <- &token{topLoc, tok_BANG, cur, nil}
		case '/':
			var rx string
			rx, err = readDelimited(&rdr, '/')
			ch <- &token{topLoc, tok_RX, cur, []string{rx}}
		case '$':
			ch <- &token{topLoc, tok_DOLLAR, cur, nil}
		case ':':
			var label string
			label, err = readIdentifier(&rdr)
			ch <- &token{topLoc, tok_LABEL, cur, []string{label}}
		case 'b', 't': // branches...
			var label string
			label, err = readIdentifier(&rdr)
			ch <- &token{topLoc, tok_CMD, cur, []string{label}}
		case 's': // substitution
			var args []string
			args, err = readSubstitution(&rdr)
			ch <- &token{topLoc, tok_CMD, cur, args}
		case 'y': // translation
			var args []string
			args, err = readTranslation(&rdr)
			ch <- &token{topLoc, tok_CMD, cur, args}
		case 'c': // change
			var txt string
			txt, err = readMultiLine(&rdr)
			ch <- &token{topLoc, tok_CHANGE, cur, []string{txt}}
		case 'i', 'a': // insert or append
			var txt string
			txt, err = readMultiLine(&rdr)
			ch <- &token{topLoc, tok_CMD, cur, []string{txt}}
		case 'r', 'w':
			var fname string
			fname, err = readIdentifier(&rdr)
			ch <- &token{topLoc, tok_CMD, cur, []string{fname}}
		default:
			if unicode.IsDigit(cur) {
				var num string
				num, err = readNumber(&rdr, cur)
				ch <- &token{topLoc, tok_NUM, cur, []string{num}}
			} else {
				// it's just a argument-free command
				ch <- &token{topLoc, tok_CMD, cur, nil}
			}
		}
	}

	if err != io.EOF {
		errch <- fmt.Errorf("Error reading... <%s> %v", err.Error(), &topLoc)
	}
}
