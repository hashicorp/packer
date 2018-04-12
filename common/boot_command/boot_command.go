package bootcommand

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var g = &grammar{
	rules: []*rule{
		{
			name: "Input",
			pos:  position{line: 6, col: 1, offset: 26},
			expr: &actionExpr{
				pos: position{line: 6, col: 10, offset: 35},
				run: (*parser).callonInput1,
				expr: &seqExpr{
					pos: position{line: 6, col: 10, offset: 35},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 6, col: 10, offset: 35},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 6, col: 15, offset: 40},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 6, col: 20, offset: 45},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 10, col: 1, offset: 75},
			expr: &actionExpr{
				pos: position{line: 10, col: 9, offset: 83},
				run: (*parser).callonExpr1,
				expr: &labeledExpr{
					pos:   position{line: 10, col: 9, offset: 83},
					label: "l",
					expr: &oneOrMoreExpr{
						pos: position{line: 10, col: 11, offset: 85},
						expr: &choiceExpr{
							pos: position{line: 10, col: 13, offset: 87},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 10, col: 13, offset: 87},
									name: "Wait",
								},
								&ruleRefExpr{
									pos:  position{line: 10, col: 20, offset: 94},
									name: "CharToggle",
								},
								&ruleRefExpr{
									pos:  position{line: 10, col: 33, offset: 107},
									name: "Special",
								},
								&ruleRefExpr{
									pos:  position{line: 10, col: 43, offset: 117},
									name: "Literal",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Wait",
			pos:  position{line: 14, col: 1, offset: 150},
			expr: &actionExpr{
				pos: position{line: 14, col: 8, offset: 157},
				run: (*parser).callonWait1,
				expr: &seqExpr{
					pos: position{line: 14, col: 8, offset: 157},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 14, col: 8, offset: 157},
							name: "ExprStart",
						},
						&litMatcher{
							pos:        position{line: 14, col: 18, offset: 167},
							val:        "wait",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 14, col: 25, offset: 174},
							label: "duration",
							expr: &zeroOrOneExpr{
								pos: position{line: 14, col: 34, offset: 183},
								expr: &choiceExpr{
									pos: position{line: 14, col: 36, offset: 185},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 14, col: 36, offset: 185},
											name: "Duration",
										},
										&ruleRefExpr{
											pos:  position{line: 14, col: 47, offset: 196},
											name: "Integer",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 14, col: 58, offset: 207},
							name: "ExprEnd",
						},
					},
				},
			},
		},
		{
			name: "CharToggle",
			pos:  position{line: 27, col: 1, offset: 453},
			expr: &actionExpr{
				pos: position{line: 27, col: 14, offset: 466},
				run: (*parser).callonCharToggle1,
				expr: &seqExpr{
					pos: position{line: 27, col: 14, offset: 466},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 27, col: 14, offset: 466},
							name: "ExprStart",
						},
						&labeledExpr{
							pos:   position{line: 27, col: 24, offset: 476},
							label: "lit",
							expr: &ruleRefExpr{
								pos:  position{line: 27, col: 29, offset: 481},
								name: "Literal",
							},
						},
						&labeledExpr{
							pos:   position{line: 27, col: 38, offset: 490},
							label: "t",
							expr: &choiceExpr{
								pos: position{line: 27, col: 41, offset: 493},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 27, col: 41, offset: 493},
										name: "On",
									},
									&ruleRefExpr{
										pos:  position{line: 27, col: 46, offset: 498},
										name: "Off",
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 27, col: 51, offset: 503},
							name: "ExprEnd",
						},
					},
				},
			},
		},
		{
			name: "Special",
			pos:  position{line: 31, col: 1, offset: 574},
			expr: &actionExpr{
				pos: position{line: 31, col: 11, offset: 584},
				run: (*parser).callonSpecial1,
				expr: &seqExpr{
					pos: position{line: 31, col: 11, offset: 584},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 31, col: 11, offset: 584},
							name: "ExprStart",
						},
						&labeledExpr{
							pos:   position{line: 31, col: 21, offset: 594},
							label: "s",
							expr: &ruleRefExpr{
								pos:  position{line: 31, col: 24, offset: 597},
								name: "SpecialKey",
							},
						},
						&labeledExpr{
							pos:   position{line: 31, col: 36, offset: 609},
							label: "t",
							expr: &zeroOrOneExpr{
								pos: position{line: 31, col: 38, offset: 611},
								expr: &choiceExpr{
									pos: position{line: 31, col: 39, offset: 612},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 31, col: 39, offset: 612},
											name: "On",
										},
										&ruleRefExpr{
											pos:  position{line: 31, col: 44, offset: 617},
											name: "Off",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 31, col: 50, offset: 623},
							name: "ExprEnd",
						},
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 38, col: 1, offset: 799},
			expr: &actionExpr{
				pos: position{line: 38, col: 10, offset: 808},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 38, col: 10, offset: 808},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 38, col: 10, offset: 808},
							expr: &litMatcher{
								pos:        position{line: 38, col: 10, offset: 808},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 38, col: 15, offset: 813},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 38, col: 23, offset: 821},
							expr: &seqExpr{
								pos: position{line: 38, col: 25, offset: 823},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 38, col: 25, offset: 823},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 38, col: 29, offset: 827},
										expr: &ruleRefExpr{
											pos:  position{line: 38, col: 29, offset: 827},
											name: "Digit",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Integer",
			pos:  position{line: 42, col: 1, offset: 873},
			expr: &choiceExpr{
				pos: position{line: 42, col: 11, offset: 883},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 42, col: 11, offset: 883},
						val:        "0",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 42, col: 17, offset: 889},
						run: (*parser).callonInteger3,
						expr: &seqExpr{
							pos: position{line: 42, col: 17, offset: 889},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 42, col: 17, offset: 889},
									name: "NonZeroDigit",
								},
								&zeroOrMoreExpr{
									pos: position{line: 42, col: 30, offset: 902},
									expr: &ruleRefExpr{
										pos:  position{line: 42, col: 30, offset: 902},
										name: "Digit",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Duration",
			pos:  position{line: 46, col: 1, offset: 966},
			expr: &actionExpr{
				pos: position{line: 46, col: 12, offset: 977},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 46, col: 12, offset: 977},
					expr: &seqExpr{
						pos: position{line: 46, col: 14, offset: 979},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 46, col: 14, offset: 979},
								name: "Number",
							},
							&ruleRefExpr{
								pos:  position{line: 46, col: 21, offset: 986},
								name: "TimeUnit",
							},
						},
					},
				},
			},
		},
		{
			name: "On",
			pos:  position{line: 50, col: 1, offset: 1049},
			expr: &actionExpr{
				pos: position{line: 50, col: 6, offset: 1054},
				run: (*parser).callonOn1,
				expr: &litMatcher{
					pos:        position{line: 50, col: 6, offset: 1054},
					val:        "on",
					ignoreCase: true,
				},
			},
		},
		{
			name: "Off",
			pos:  position{line: 54, col: 1, offset: 1087},
			expr: &actionExpr{
				pos: position{line: 54, col: 7, offset: 1093},
				run: (*parser).callonOff1,
				expr: &litMatcher{
					pos:        position{line: 54, col: 7, offset: 1093},
					val:        "off",
					ignoreCase: true,
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 58, col: 1, offset: 1128},
			expr: &actionExpr{
				pos: position{line: 58, col: 11, offset: 1138},
				run: (*parser).callonLiteral1,
				expr: &anyMatcher{
					line: 58, col: 11, offset: 1138,
				},
			},
		},
		{
			name: "ExprEnd",
			pos:  position{line: 63, col: 1, offset: 1219},
			expr: &litMatcher{
				pos:        position{line: 63, col: 11, offset: 1229},
				val:        ">",
				ignoreCase: false,
			},
		},
		{
			name: "ExprStart",
			pos:  position{line: 64, col: 1, offset: 1233},
			expr: &litMatcher{
				pos:        position{line: 64, col: 13, offset: 1245},
				val:        "<",
				ignoreCase: false,
			},
		},
		{
			name: "SpecialKey",
			pos:  position{line: 65, col: 1, offset: 1249},
			expr: &choiceExpr{
				pos: position{line: 65, col: 14, offset: 1262},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 65, col: 14, offset: 1262},
						val:        "bs",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 65, col: 22, offset: 1270},
						val:        "del",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 65, col: 31, offset: 1279},
						val:        "enter",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 65, col: 42, offset: 1290},
						val:        "esc",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 65, col: 51, offset: 1299},
						val:        "f10",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 65, col: 60, offset: 1308},
						val:        "f11",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 65, col: 69, offset: 1317},
						val:        "f12",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 66, col: 11, offset: 1334},
						val:        "f1",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 66, col: 19, offset: 1342},
						val:        "f2",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 66, col: 27, offset: 1350},
						val:        "f3",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 66, col: 35, offset: 1358},
						val:        "f4",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 66, col: 43, offset: 1366},
						val:        "f5",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 66, col: 51, offset: 1374},
						val:        "f6",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 66, col: 59, offset: 1382},
						val:        "f7",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 66, col: 67, offset: 1390},
						val:        "f8",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 66, col: 75, offset: 1398},
						val:        "f9",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 67, col: 12, offset: 1415},
						val:        "return",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 67, col: 24, offset: 1427},
						val:        "tab",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 67, col: 33, offset: 1436},
						val:        "up",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 67, col: 41, offset: 1444},
						val:        "down",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 67, col: 51, offset: 1454},
						val:        "spacebar",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 67, col: 65, offset: 1468},
						val:        "insert",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 67, col: 77, offset: 1480},
						val:        "home",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 68, col: 11, offset: 1498},
						val:        "end",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 68, col: 20, offset: 1507},
						val:        "pageup",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 68, col: 32, offset: 1519},
						val:        "pagedown",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 68, col: 46, offset: 1533},
						val:        "leftalt",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 68, col: 59, offset: 1546},
						val:        "leftctrl",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 68, col: 73, offset: 1560},
						val:        "leftshift",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 69, col: 11, offset: 1583},
						val:        "rightalt",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 69, col: 25, offset: 1597},
						val:        "rightctrl",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 69, col: 40, offset: 1612},
						val:        "rightshift",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 69, col: 56, offset: 1628},
						val:        "leftsuper",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 69, col: 71, offset: 1643},
						val:        "rightsuper",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 70, col: 11, offset: 1667},
						val:        "left",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 70, col: 21, offset: 1677},
						val:        "right",
						ignoreCase: true,
					},
				},
			},
		},
		{
			name: "NonZeroDigit",
			pos:  position{line: 72, col: 1, offset: 1687},
			expr: &charClassMatcher{
				pos:        position{line: 72, col: 16, offset: 1702},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 73, col: 1, offset: 1708},
			expr: &charClassMatcher{
				pos:        position{line: 73, col: 9, offset: 1716},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "TimeUnit",
			pos:  position{line: 74, col: 1, offset: 1722},
			expr: &choiceExpr{
				pos: position{line: 74, col: 13, offset: 1734},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 74, col: 13, offset: 1734},
						val:        "ns",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 74, col: 20, offset: 1741},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 74, col: 27, offset: 1748},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 74, col: 34, offset: 1756},
						val:        "ms",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 74, col: 41, offset: 1763},
						val:        "s",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 74, col: 47, offset: 1769},
						val:        "m",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 74, col: 53, offset: 1775},
						val:        "h",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "_",
			displayName: "\"whitespace\"",
			pos:         position{line: 76, col: 1, offset: 1781},
			expr: &zeroOrMoreExpr{
				pos: position{line: 76, col: 19, offset: 1799},
				expr: &charClassMatcher{
					pos:        position{line: 76, col: 19, offset: 1799},
					val:        "[ \\n\\t\\r]",
					chars:      []rune{' ', '\n', '\t', '\r'},
					ignoreCase: false,
					inverted:   false,
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 78, col: 1, offset: 1811},
			expr: &notExpr{
				pos: position{line: 78, col: 8, offset: 1818},
				expr: &anyMatcher{
					line: 78, col: 9, offset: 1819,
				},
			},
		},
	},
}

func (c *current) onInput1(expr interface{}) (interface{}, error) {
	return expr, nil
}

func (p *parser) callonInput1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInput1(stack["expr"])
}

func (c *current) onExpr1(l interface{}) (interface{}, error) {
	return l, nil
}

func (p *parser) callonExpr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onExpr1(stack["l"])
}

func (c *current) onWait1(duration interface{}) (interface{}, error) {
	var d time.Duration
	switch t := duration.(type) {
	case time.Duration:
		d = t
	case int64:
		d = time.Duration(t) * time.Second
	default:
		d = time.Second
	}
	return &waitExpression{d}, nil
}

func (p *parser) callonWait1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onWait1(stack["duration"])
}

func (c *current) onCharToggle1(lit, t interface{}) (interface{}, error) {
	return &literal{lit.(*literal).s, t.(KeyAction)}, nil
}

func (p *parser) callonCharToggle1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCharToggle1(stack["lit"], stack["t"])
}

func (c *current) onSpecial1(s, t interface{}) (interface{}, error) {
	if t == nil {
		return &specialExpression{string(s.([]byte)), KeyPress}, nil
	}
	return &specialExpression{string(s.([]byte)), t.(KeyAction)}, nil
}

func (p *parser) callonSpecial1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSpecial1(stack["s"], stack["t"])
}

func (c *current) onNumber1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonNumber1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNumber1()
}

func (c *current) onInteger3() (interface{}, error) {
	return strconv.ParseInt(string(c.text), 10, 64)
}

func (p *parser) callonInteger3() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInteger3()
}

func (c *current) onDuration1() (interface{}, error) {
	return time.ParseDuration(string(c.text))
}

func (p *parser) callonDuration1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDuration1()
}

func (c *current) onOn1() (interface{}, error) {
	return KeyOn, nil
}

func (p *parser) callonOn1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onOn1()
}

func (c *current) onOff1() (interface{}, error) {
	return KeyOff, nil
}

func (p *parser) callonOff1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onOff1()
}

func (c *current) onLiteral1() (interface{}, error) {
	r, _ := utf8.DecodeRune(c.text)
	return &literal{r, KeyPress}, nil
}

func (p *parser) callonLiteral1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLiteral1()
}

var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEntrypoint is returned when the specified entrypoint rule
	// does not exit.
	errInvalidEntrypoint = errors.New("invalid entrypoint")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errMaxExprCnt is used to signal that the maximum number of
	// expressions have been parsed.
	errMaxExprCnt = errors.New("max number of expresssions parsed")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// MaxExpressions creates an Option to stop parsing after the provided
// number of expressions have been parsed, if the value is 0 then the parser will
// parse for as many steps as needed (possibly an infinite number).
//
// The default for maxExprCnt is 0.
func MaxExpressions(maxExprCnt uint64) Option {
	return func(p *parser) Option {
		oldMaxExprCnt := p.maxExprCnt
		p.maxExprCnt = maxExprCnt
		return MaxExpressions(oldMaxExprCnt)
	}
}

// Entrypoint creates an Option to set the rule name to use as entrypoint.
// The rule name must have been specified in the -alternate-entrypoints
// if generating the parser with the -optimize-grammar flag, otherwise
// it may have been optimized out. Passing an empty string sets the
// entrypoint to the first rule in the grammar.
//
// The default is to start parsing at the first rule in the grammar.
func Entrypoint(ruleName string) Option {
	return func(p *parser) Option {
		oldEntrypoint := p.entrypoint
		p.entrypoint = ruleName
		if ruleName == "" {
			p.entrypoint = g.rules[0].name
		}
		return Entrypoint(oldEntrypoint)
	}
}

// Statistics adds a user provided Stats struct to the parser to allow
// the user to process the results after the parsing has finished.
// Also the key for the "no match" counter is set.
//
// Example usage:
//
//     input := "input"
//     stats := Stats{}
//     _, err := Parse("input-file", []byte(input), Statistics(&stats, "no match"))
//     if err != nil {
//         log.Panicln(err)
//     }
//     b, err := json.MarshalIndent(stats.ChoiceAltCnt, "", "  ")
//     if err != nil {
//         log.Panicln(err)
//     }
//     fmt.Println(string(b))
//
func Statistics(stats *Stats, choiceNoMatch string) Option {
	return func(p *parser) Option {
		oldStats := p.Stats
		p.Stats = stats
		oldChoiceNoMatch := p.choiceNoMatch
		p.choiceNoMatch = choiceNoMatch
		if p.Stats.ChoiceAltCnt == nil {
			p.Stats.ChoiceAltCnt = make(map[string]map[string]int)
		}
		return Statistics(oldStats, oldChoiceNoMatch)
	}
}

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// AllowInvalidUTF8 creates an Option to allow invalid UTF-8 bytes.
// Every invalid UTF-8 byte is treated as a utf8.RuneError (U+FFFD)
// by character class matchers and is matched by the any matcher.
// The returned matched value, c.text and c.offset are NOT affected.
//
// The default is false.
func AllowInvalidUTF8(b bool) Option {
	return func(p *parser) Option {
		old := p.allowInvalidUTF8
		p.allowInvalidUTF8 = b
		return AllowInvalidUTF8(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// GlobalStore creates an Option to set a key to a certain value in
// the globalStore.
func GlobalStore(key string, value interface{}) Option {
	return func(p *parser) Option {
		old := p.cur.globalStore[key]
		p.cur.globalStore[key] = value
		return GlobalStore(key, old)
	}
}

// InitState creates an Option to set a key to a certain value in
// the global "state" store.
func InitState(key string, value interface{}) Option {
	return func(p *parser) Option {
		old := p.cur.state[key]
		p.cur.state[key] = value
		return InitState(key, old)
	}
}

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (i interface{}, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = closeErr
		}
	}()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (interface{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (interface{}, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d [%d]", p.line, p.col, p.offset)
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match

	// state is a store for arbitrary key,value pairs that the user wants to be
	// tied to the backtracking of the parser.
	// This is always rolled back if a parsing rule fails.
	state storeDict

	// globalStore is a general store for the user to store arbitrary key-value
	// pairs that they need to manage and that they do not want tied to the
	// backtracking of the parser. This is only modified by the user and never
	// rolled back by the parser. It is always up to the user to keep this in a
	// consistent state.
	globalStore storeDict
}

type storeDict map[string]interface{}

// the AST types...

type grammar struct {
	pos   position
	rules []*rule
}

type rule struct {
	pos         position
	name        string
	displayName string
	expr        interface{}
}

type choiceExpr struct {
	pos          position
	alternatives []interface{}
}

type actionExpr struct {
	pos  position
	expr interface{}
	run  func(*parser) (interface{}, error)
}

type recoveryExpr struct {
	pos          position
	expr         interface{}
	recoverExpr  interface{}
	failureLabel []string
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type throwExpr struct {
	pos   position
	label string
}

type labeledExpr struct {
	pos   position
	label string
	expr  interface{}
}

type expr struct {
	pos  position
	expr interface{}
}

type andExpr expr
type notExpr expr
type zeroOrOneExpr expr
type zeroOrMoreExpr expr
type oneOrMoreExpr expr

type ruleRefExpr struct {
	pos  position
	name string
}

type stateCodeExpr struct {
	pos position
	run func(*parser) error
}

type andCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type notCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
}

type charClassMatcher struct {
	pos             position
	val             string
	basicLatinChars [128]bool
	chars           []rune
	ranges          []rune
	classes         []*unicode.RangeTable
	ignoreCase      bool
	inverted        bool
}

type anyMatcher position

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner    error
	pos      position
	prefix   string
	expected []string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	stats := Stats{
		ChoiceAltCnt: make(map[string]map[string]int),
	}

	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
		cur: current{
			state:       make(storeDict),
			globalStore: make(storeDict),
		},
		maxFailPos:      position{col: 1, line: 1},
		maxFailExpected: make([]string, 0, 20),
		Stats:           &stats,
		// start rule is rule [0] unless an alternate entrypoint is specified
		entrypoint: g.rules[0].name,
		emptyState: make(storeDict),
	}
	p.setOptions(opts)

	if p.maxExprCnt == 0 {
		p.maxExprCnt = math.MaxUint64
	}

	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

type resultTuple struct {
	v   interface{}
	b   bool
	end savepoint
}

const choiceNoMatch = -1

// Stats stores some statistics, gathered during parsing
type Stats struct {
	// ExprCnt counts the number of expressions processed during parsing
	// This value is compared to the maximum number of expressions allowed
	// (set by the MaxExpressions option).
	ExprCnt uint64

	// ChoiceAltCnt is used to count for each ordered choice expression,
	// which alternative is used how may times.
	// These numbers allow to optimize the order of the ordered choice expression
	// to increase the performance of the parser
	//
	// The outer key of ChoiceAltCnt is composed of the name of the rule as well
	// as the line and the column of the ordered choice.
	// The inner key of ChoiceAltCnt is the number (one-based) of the matching alternative.
	// For each alternative the number of matches are counted. If an ordered choice does not
	// match, a special counter is incremented. The name of this counter is set with
	// the parser option Statistics.
	// For an alternative to be included in ChoiceAltCnt, it has to match at least once.
	ChoiceAltCnt map[string]map[string]int
}

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	depth   int
	recover bool
	debug   bool

	memoize bool
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[interface{}]resultTuple

	// rules table, maps the rule identifier to the rule node
	rules map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]interface{}
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// parse fail
	maxFailPos            position
	maxFailExpected       []string
	maxFailInvertExpected bool

	// max number of expressions to be parsed
	maxExprCnt uint64
	// entrypoint for the parser
	entrypoint string

	allowInvalidUTF8 bool

	*Stats

	choiceNoMatch string
	// recovery expression stack, keeps track of the currently available recovery expression, these are traversed in reverse
	recoveryStack []map[string]interface{}

	// emptyState contains an empty storeDict, which is used to optimize cloneState if global "state" store is not used.
	emptyState storeDict
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]interface{})
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

// push a recovery expression with its labels to the recoveryStack
func (p *parser) pushRecovery(labels []string, expr interface{}) {
	if cap(p.recoveryStack) == len(p.recoveryStack) {
		// create new empty slot in the stack
		p.recoveryStack = append(p.recoveryStack, nil)
	} else {
		// slice to 1 more
		p.recoveryStack = p.recoveryStack[:len(p.recoveryStack)+1]
	}

	m := make(map[string]interface{}, len(labels))
	for _, fl := range labels {
		m[fl] = expr
	}
	p.recoveryStack[len(p.recoveryStack)-1] = m
}

// pop a recovery expression from the recoveryStack
func (p *parser) popRecovery() {
	// GC that map
	p.recoveryStack[len(p.recoveryStack)-1] = nil

	p.recoveryStack = p.recoveryStack[:len(p.recoveryStack)-1]
}

func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) in(s string) string {
	p.depth++
	return p.print(strings.Repeat(" ", p.depth)+">", s)
}

func (p *parser) out(s string) string {
	p.depth--
	return p.print(strings.Repeat(" ", p.depth)+"<", s)
}

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position, []string{})
}

func (p *parser) addErrAt(err error, pos position, expected []string) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String(), expected: expected}
	p.errs.add(pe)
}

func (p *parser) failAt(fail bool, pos position, want string) {
	// process fail if parsing fails and not inverted or parsing succeeds and invert is set
	if fail == p.maxFailInvertExpected {
		if pos.offset < p.maxFailPos.offset {
			return
		}

		if pos.offset > p.maxFailPos.offset {
			p.maxFailPos = pos
			p.maxFailExpected = p.maxFailExpected[:0]
		}

		if p.maxFailInvertExpected {
			want = "!" + want
		}
		p.maxFailExpected = append(p.maxFailExpected, want)
	}
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError && n == 1 { // see utf8.DecodeRune
		if !p.allowInvalidUTF8 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	if p.debug {
		defer p.out(p.in("restore"))
	}
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// Cloner is implemented by any value that has a Clone method, which returns a
// copy of the value. This is mainly used for types which are not passed by
// value (e.g map, slice, chan) or structs that contain such types.
//
// This is used in conjunction with the global state feature to create proper
// copies of the state to allow the parser to properly restore the state in
// the case of backtracking.
type Cloner interface {
	Clone() interface{}
}

// clone and return parser current state.
func (p *parser) cloneState() storeDict {
	if p.debug {
		defer p.out(p.in("cloneState"))
	}

	if len(p.cur.state) == 0 {
		if len(p.emptyState) > 0 {
			p.emptyState = make(storeDict)
		}
		return p.emptyState
	}

	state := make(storeDict, len(p.cur.state))
	for k, v := range p.cur.state {
		if c, ok := v.(Cloner); ok {
			state[k] = c.Clone()
		} else {
			state[k] = v
		}
	}
	return state
}

// restore parser current state to the state storeDict.
// every restoreState should applied only one time for every cloned state
func (p *parser) restoreState(state storeDict) {
	if p.debug {
		defer p.out(p.in("restoreState"))
	}
	p.cur.state = state
}

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

func (p *parser) getMemoized(node interface{}) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node interface{}, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[interface{}]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[interface{}]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

func (p *parser) parse(g *grammar) (val interface{}, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	startRule, ok := p.rules[p.entrypoint]
	if !ok {
		p.addErr(errInvalidEntrypoint)
		return nil, p.errs.err()
	}

	p.read() // advance to first rune
	val, ok = p.parseRule(startRule)
	if !ok {
		if len(*p.errs) == 0 {
			// If parsing fails, but no errors have been recorded, the expected values
			// for the farthest parser position are returned as error.
			maxFailExpectedMap := make(map[string]struct{}, len(p.maxFailExpected))
			for _, v := range p.maxFailExpected {
				maxFailExpectedMap[v] = struct{}{}
			}
			expected := make([]string, 0, len(maxFailExpectedMap))
			eof := false
			if _, ok := maxFailExpectedMap["!."]; ok {
				delete(maxFailExpectedMap, "!.")
				eof = true
			}
			for k := range maxFailExpectedMap {
				expected = append(expected, k)
			}
			sort.Strings(expected)
			if eof {
				expected = append(expected, "EOF")
			}
			p.addErrAt(errors.New("no match found, expected: "+listJoin(expected, ", ", "or")), p.maxFailPos, expected)
		}

		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func listJoin(list []string, sep string, lastSep string) string {
	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	default:
		return fmt.Sprintf("%s %s %s", strings.Join(list[:len(list)-1], sep), lastSep, list[len(list)-1])
	}
}

func (p *parser) parseRule(rule *rule) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}

	if p.memoize {
		res, ok := p.getMemoized(rule)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
	}

	start := p.pt
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExpr(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}

	if p.memoize {
		p.setMemoized(start, rule, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseExpr(expr interface{}) (interface{}, bool) {
	var pt savepoint

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.ExprCnt++
	if p.ExprCnt > p.maxExprCnt {
		panic(errMaxExprCnt)
	}

	var val interface{}
	var ok bool
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andCodeExpr:
		val, ok = p.parseAndCodeExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notCodeExpr:
		val, ok = p.parseNotCodeExpr(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *recoveryExpr:
		val, ok = p.parseRecoveryExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *stateCodeExpr:
		val, ok = p.parseStateCodeExpr(expr)
	case *throwExpr:
		val, ok = p.parseThrowExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	if p.memoize {
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	start := p.pt
	val, ok := p.parseExpr(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		state := p.cloneState()
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position, []string{})
		}
		p.restoreState(state)

		val = actVal
	}
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}
	return val, ok
}

func (p *parser) parseAndCodeExpr(and *andCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndCodeExpr"))
	}

	state := p.cloneState()

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	p.restoreState(state)

	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	state := p.cloneState()
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restoreState(state)
	p.restore(pt)

	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn == utf8.RuneError && p.pt.w == 0 {
		// EOF - see utf8.DecodeRune
		p.failAt(false, p.pt.position, ".")
		return nil, false
	}
	start := p.pt
	p.read()
	p.failAt(true, start.position, ".")
	return p.sliceFrom(start), true
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	start := p.pt

	// can't match EOF
	if cur == utf8.RuneError && p.pt.w == 0 { // see utf8.DecodeRune
		p.failAt(false, start.position, chr.val)
		return nil, false
	}

	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		p.failAt(true, start.position, chr.val)
		return p.sliceFrom(start), true
	}
	p.failAt(false, start.position, chr.val)
	return nil, false
}

func (p *parser) incChoiceAltCnt(ch *choiceExpr, altI int) {
	choiceIdent := fmt.Sprintf("%s %d:%d", p.rstack[len(p.rstack)-1].name, ch.pos.line, ch.pos.col)
	m := p.ChoiceAltCnt[choiceIdent]
	if m == nil {
		m = make(map[string]int)
		p.ChoiceAltCnt[choiceIdent] = m
	}
	// We increment altI by 1, so the keys do not start at 0
	alt := strconv.Itoa(altI + 1)
	if altI == choiceNoMatch {
		alt = p.choiceNoMatch
	}
	m[alt]++
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for altI, alt := range ch.alternatives {
		// dummy assignment to prevent compile error if optimized
		_ = altI

		state := p.cloneState()

		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			p.incChoiceAltCnt(ch, altI)
			return val, ok
		}
		p.restoreState(state)
	}
	p.incChoiceAltCnt(ch, choiceNoMatch)
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	p.pushV()
	val, ok := p.parseExpr(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseLitMatcher(lit *litMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	ignoreCase := ""
	if lit.ignoreCase {
		ignoreCase = "i"
	}
	val := fmt.Sprintf("%q%s", lit.val, ignoreCase)
	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.failAt(false, start.position, val)
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	p.failAt(true, start.position, val)
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	state := p.cloneState()

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	p.restoreState(state)

	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	state := p.cloneState()
	p.pushV()
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	_, ok := p.parseExpr(not.expr)
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	p.popV()
	p.restoreState(state)
	p.restore(pt)

	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRecoveryExpr(recover *recoveryExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRecoveryExpr (" + strings.Join(recover.failureLabel, ",") + ")"))
	}

	p.pushRecovery(recover.failureLabel, recover.recoverExpr)
	val, ok := p.parseExpr(recover.expr)
	p.popRecovery()

	return val, ok
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRule(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	vals := make([]interface{}, 0, len(seq.exprs))

	pt := p.pt
	state := p.cloneState()
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restoreState(state)
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseStateCodeExpr(state *stateCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseStateCodeExpr"))
	}

	err := state.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, true
}

func (p *parser) parseThrowExpr(expr *throwExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseThrowExpr"))
	}

	for i := len(p.recoveryStack) - 1; i >= 0; i-- {
		if recoverExpr, ok := p.recoveryStack[i][expr.label]; ok {
			if val, ok := p.parseExpr(recoverExpr); ok {
				return val, ok
			}
		}
	}

	return nil, false
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	p.pushV()
	val, _ := p.parseExpr(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}
