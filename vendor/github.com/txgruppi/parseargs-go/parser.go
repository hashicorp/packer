package parseargs

import (
	"errors"
	"regexp"
	"strings"
)

var (
	whitespaceRegexp       = regexp.MustCompile("\\s")
	specialCharsRegexp     = regexp.MustCompile(`\s|"|'`)
	backSlashRemovalRegexp = regexp.MustCompile(`\\([\s"'\\])`)

	// ErrInvalidArgument is the error returned when an unexpected character
	// is found by the parser.
	ErrInvalidArgument = errors.New("invalid argument(s)")

	// ErrInvalidSyntax is the error returned when some of the syntax rules are
	// violeted by the input.
	ErrInvalidSyntax = errors.New("invalid syntax")

	// ErrUnexpectedEndOfInput is the error returned when the parser gets to the
	// end of the string with an unfinished string.
	ErrUnexpectedEndOfInput = errors.New("unexpected end of input")
)

// Parse parses a string into a list or arguments. The default argument
// separator is one or a sequence of whitespaces but it also understands
// quotted string and escaped quotes.
func Parse(input string) ([]string, error) {
	return newParser(input).parse()
}

func newParser(input string) *parser {
	runes := []rune(strings.TrimSpace(input))
	return &parser{
		runes:  runes,
		length: len(runes),
	}
}

type parser struct {
	runes      []rune
	length     int
	reading    bool
	startChar  rune
	startIndex int
}

func (p *parser) parse() ([]string, error) {
	result := []string{}
	for index, current := range p.runes {
		if p.checkInvalidArgument(current) {
			return nil, ErrInvalidArgument
		}

		if p.shouldStartReadingWord(current) {
			p.reading = true
			p.startChar = ' '
			p.startIndex = index

			if p.shouldFinishReadingAtEndOfInput(index) {
				result = append(result, string(p.read(p.startIndex, p.length)))
			}
			continue
		}

		if p.shouldStartReadingQuottedString(current) {
			p.reading = true
			p.startChar = current
			p.startIndex = index
			continue
		}

		if !p.reading {
			continue
		}

		if p.shouldFinishReadingWord(current) {
			if !p.hasValidBackslash(index) {
				return nil, ErrInvalidSyntax
			}
			result = append(result, string(p.read(p.startIndex, index)))
			continue
		}

		if p.shouldFinishReadingQuottedString(index, current) {
			result = append(result, string(p.read(p.startIndex+1, index)))
			continue
		}

		if p.shouldFinishReadingAtEndOfInput(index) {
			result = append(result, string(p.read(p.startIndex, p.length)))
			continue
		}
	}

	if p.hasEndedUnexpectedly() {
		return nil, ErrUnexpectedEndOfInput
	}

	return p.cleanUpResult(result), nil
}

func (p *parser) shouldFinishReadingAtEndOfInput(index int) bool {
	return p.isEndOfInput(index) && p.startChar == ' '
}

func (p *parser) cleanUpResult(result []string) []string {
	for index, value := range result {
		result[index] = backSlashRemovalRegexp.ReplaceAllString(value, "$1")
	}
	return result
}

func (p *parser) hasEndedUnexpectedly() bool {
	return p.startIndex >= 0 || p.startChar != 0
}

func (p *parser) shouldFinishReadingQuottedString(index int, char rune) bool {
	return p.startChar == char && p.isSpecial(p.startChar) && p.hasValidBackslash(index)
}

func (p *parser) shouldFinishReadingWord(char rune) bool {
	return p.startChar == ' ' && p.isWhitespace(char)
}

func (p *parser) shouldStartReadingQuottedString(char rune) bool {
	return !p.reading && p.isSpecial(char) && !p.isWhitespace(char)
}

func (p *parser) isEndOfInput(index int) bool {
	return index == p.length-1
}

func (p *parser) shouldStartReadingWord(char rune) bool {
	return !(p.reading || p.isSpecial(char))
}

func (p *parser) checkInvalidArgument(char rune) bool {
	return p.reading && p.startChar == ' ' && p.isSpecial(char) && !p.isWhitespace(char)
}

func (p *parser) read(start, end int) []rune {
	p.reading = false
	p.startChar = 0
	p.startIndex = -1
	return p.runes[start:end]
}

func (p *parser) isWhitespace(r rune) bool {
	return whitespaceRegexp.MatchString(string(r))
}

func (p *parser) isSpecial(r rune) bool {
	return specialCharsRegexp.MatchString(string(r))
}

func (p *parser) hasValidBackslash(index int) bool {
	counter := 0

	for {
		if index-1-counter < 0 {
			break
		}

		if p.runes[index-1-counter] == '\\' {
			counter++
			continue
		}

		break
	}

	return counter%2 == 0
}
