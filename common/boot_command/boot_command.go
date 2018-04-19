package bootcommand

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

/*
func main() {
    in := "<wait><wait10><wait1s><wait1m2ns>"
    in += "foo/bar > one"
    in += "<fOn> b<fOff>"
    in += "<f3><f12><spacebar><leftalt><rightshift><rightsuper>"
    got, err := ParseReader("", strings.NewReader(in))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s\n", got)
}
*/

var g = &grammar{
	rules: []*rule{
		{
			name: "Input",
			pos:  position{line: 21, col: 1, offset: 345},
			expr: &actionExpr{
				pos: position{line: 21, col: 10, offset: 354},
				run: (*parser).callonInput1,
				expr: &seqExpr{
					pos: position{line: 21, col: 10, offset: 354},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 21, col: 10, offset: 354},
							label: "expr",
							expr: &ruleRefExpr{
								pos:  position{line: 21, col: 15, offset: 359},
								name: "Expr",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 21, col: 20, offset: 364},
							name: "EOF",
						},
					},
				},
			},
		},
		{
			name: "Expr",
			pos:  position{line: 25, col: 1, offset: 394},
			expr: &actionExpr{
				pos: position{line: 25, col: 9, offset: 402},
				run: (*parser).callonExpr1,
				expr: &labeledExpr{
					pos:   position{line: 25, col: 9, offset: 402},
					label: "l",
					expr: &oneOrMoreExpr{
						pos: position{line: 25, col: 11, offset: 404},
						expr: &choiceExpr{
							pos: position{line: 25, col: 13, offset: 406},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 25, col: 13, offset: 406},
									name: "Wait",
								},
								&ruleRefExpr{
									pos:  position{line: 25, col: 20, offset: 413},
									name: "CharToggle",
								},
								&ruleRefExpr{
									pos:  position{line: 25, col: 33, offset: 426},
									name: "Special",
								},
								&ruleRefExpr{
									pos:  position{line: 25, col: 43, offset: 436},
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
			pos:  position{line: 29, col: 1, offset: 469},
			expr: &actionExpr{
				pos: position{line: 29, col: 8, offset: 476},
				run: (*parser).callonWait1,
				expr: &seqExpr{
					pos: position{line: 29, col: 8, offset: 476},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 29, col: 8, offset: 476},
							name: "ExprStart",
						},
						&litMatcher{
							pos:        position{line: 29, col: 18, offset: 486},
							val:        "wait",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 29, col: 25, offset: 493},
							label: "duration",
							expr: &zeroOrOneExpr{
								pos: position{line: 29, col: 34, offset: 502},
								expr: &choiceExpr{
									pos: position{line: 29, col: 36, offset: 504},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 29, col: 36, offset: 504},
											name: "Duration",
										},
										&ruleRefExpr{
											pos:  position{line: 29, col: 47, offset: 515},
											name: "Integer",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 29, col: 58, offset: 526},
							name: "ExprEnd",
						},
					},
				},
			},
		},
		{
			name: "CharToggle",
			pos:  position{line: 42, col: 1, offset: 772},
			expr: &actionExpr{
				pos: position{line: 42, col: 14, offset: 785},
				run: (*parser).callonCharToggle1,
				expr: &seqExpr{
					pos: position{line: 42, col: 14, offset: 785},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 42, col: 14, offset: 785},
							name: "ExprStart",
						},
						&labeledExpr{
							pos:   position{line: 42, col: 24, offset: 795},
							label: "lit",
							expr: &ruleRefExpr{
								pos:  position{line: 42, col: 29, offset: 800},
								name: "Literal",
							},
						},
						&labeledExpr{
							pos:   position{line: 42, col: 38, offset: 809},
							label: "t",
							expr: &choiceExpr{
								pos: position{line: 42, col: 41, offset: 812},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 42, col: 41, offset: 812},
										name: "On",
									},
									&ruleRefExpr{
										pos:  position{line: 42, col: 46, offset: 817},
										name: "Off",
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 42, col: 51, offset: 822},
							name: "ExprEnd",
						},
					},
				},
			},
		},
		{
			name: "Special",
			pos:  position{line: 46, col: 1, offset: 893},
			expr: &actionExpr{
				pos: position{line: 46, col: 11, offset: 903},
				run: (*parser).callonSpecial1,
				expr: &seqExpr{
					pos: position{line: 46, col: 11, offset: 903},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 46, col: 11, offset: 903},
							name: "ExprStart",
						},
						&labeledExpr{
							pos:   position{line: 46, col: 21, offset: 913},
							label: "s",
							expr: &ruleRefExpr{
								pos:  position{line: 46, col: 24, offset: 916},
								name: "SpecialKey",
							},
						},
						&labeledExpr{
							pos:   position{line: 46, col: 36, offset: 928},
							label: "t",
							expr: &zeroOrOneExpr{
								pos: position{line: 46, col: 38, offset: 930},
								expr: &choiceExpr{
									pos: position{line: 46, col: 39, offset: 931},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 46, col: 39, offset: 931},
											name: "On",
										},
										&ruleRefExpr{
											pos:  position{line: 46, col: 44, offset: 936},
											name: "Off",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 46, col: 50, offset: 942},
							name: "ExprEnd",
						},
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 55, col: 1, offset: 1211},
			expr: &actionExpr{
				pos: position{line: 55, col: 10, offset: 1220},
				run: (*parser).callonNumber1,
				expr: &seqExpr{
					pos: position{line: 55, col: 10, offset: 1220},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 55, col: 10, offset: 1220},
							expr: &litMatcher{
								pos:        position{line: 55, col: 10, offset: 1220},
								val:        "-",
								ignoreCase: false,
							},
						},
						&ruleRefExpr{
							pos:  position{line: 55, col: 15, offset: 1225},
							name: "Integer",
						},
						&zeroOrOneExpr{
							pos: position{line: 55, col: 23, offset: 1233},
							expr: &seqExpr{
								pos: position{line: 55, col: 25, offset: 1235},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 55, col: 25, offset: 1235},
										val:        ".",
										ignoreCase: false,
									},
									&oneOrMoreExpr{
										pos: position{line: 55, col: 29, offset: 1239},
										expr: &ruleRefExpr{
											pos:  position{line: 55, col: 29, offset: 1239},
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
			pos:  position{line: 59, col: 1, offset: 1285},
			expr: &choiceExpr{
				pos: position{line: 59, col: 11, offset: 1295},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 59, col: 11, offset: 1295},
						val:        "0",
						ignoreCase: false,
					},
					&actionExpr{
						pos: position{line: 59, col: 17, offset: 1301},
						run: (*parser).callonInteger3,
						expr: &seqExpr{
							pos: position{line: 59, col: 17, offset: 1301},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 59, col: 17, offset: 1301},
									name: "NonZeroDigit",
								},
								&zeroOrMoreExpr{
									pos: position{line: 59, col: 30, offset: 1314},
									expr: &ruleRefExpr{
										pos:  position{line: 59, col: 30, offset: 1314},
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
			pos:  position{line: 63, col: 1, offset: 1378},
			expr: &actionExpr{
				pos: position{line: 63, col: 12, offset: 1389},
				run: (*parser).callonDuration1,
				expr: &oneOrMoreExpr{
					pos: position{line: 63, col: 12, offset: 1389},
					expr: &seqExpr{
						pos: position{line: 63, col: 14, offset: 1391},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 63, col: 14, offset: 1391},
								name: "Number",
							},
							&ruleRefExpr{
								pos:  position{line: 63, col: 21, offset: 1398},
								name: "TimeUnit",
							},
						},
					},
				},
			},
		},
		{
			name: "On",
			pos:  position{line: 67, col: 1, offset: 1461},
			expr: &actionExpr{
				pos: position{line: 67, col: 6, offset: 1466},
				run: (*parser).callonOn1,
				expr: &litMatcher{
					pos:        position{line: 67, col: 6, offset: 1466},
					val:        "on",
					ignoreCase: true,
				},
			},
		},
		{
			name: "Off",
			pos:  position{line: 71, col: 1, offset: 1499},
			expr: &actionExpr{
				pos: position{line: 71, col: 7, offset: 1505},
				run: (*parser).callonOff1,
				expr: &litMatcher{
					pos:        position{line: 71, col: 7, offset: 1505},
					val:        "off",
					ignoreCase: true,
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 75, col: 1, offset: 1540},
			expr: &actionExpr{
				pos: position{line: 75, col: 11, offset: 1550},
				run: (*parser).callonLiteral1,
				expr: &anyMatcher{
					line: 75, col: 11, offset: 1550,
				},
			},
		},
		{
			name: "ExprEnd",
			pos:  position{line: 80, col: 1, offset: 1631},
			expr: &litMatcher{
				pos:        position{line: 80, col: 11, offset: 1641},
				val:        ">",
				ignoreCase: false,
			},
		},
		{
			name: "ExprStart",
			pos:  position{line: 81, col: 1, offset: 1645},
			expr: &litMatcher{
				pos:        position{line: 81, col: 13, offset: 1657},
				val:        "<",
				ignoreCase: false,
			},
		},
		{
			name: "SpecialKey",
			pos:  position{line: 82, col: 1, offset: 1661},
			expr: &choiceExpr{
				pos: position{line: 82, col: 14, offset: 1674},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 82, col: 14, offset: 1674},
						val:        "bs",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 82, col: 22, offset: 1682},
						val:        "del",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 82, col: 31, offset: 1691},
						val:        "enter",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 82, col: 42, offset: 1702},
						val:        "esc",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 82, col: 51, offset: 1711},
						val:        "f10",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 82, col: 60, offset: 1720},
						val:        "f11",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 82, col: 69, offset: 1729},
						val:        "f12",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 83, col: 11, offset: 1746},
						val:        "f1",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 83, col: 19, offset: 1754},
						val:        "f2",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 83, col: 27, offset: 1762},
						val:        "f3",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 83, col: 35, offset: 1770},
						val:        "f4",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 83, col: 43, offset: 1778},
						val:        "f5",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 83, col: 51, offset: 1786},
						val:        "f6",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 83, col: 59, offset: 1794},
						val:        "f7",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 83, col: 67, offset: 1802},
						val:        "f8",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 83, col: 75, offset: 1810},
						val:        "f9",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 84, col: 12, offset: 1827},
						val:        "return",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 84, col: 24, offset: 1839},
						val:        "tab",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 84, col: 33, offset: 1848},
						val:        "up",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 84, col: 41, offset: 1856},
						val:        "down",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 84, col: 51, offset: 1866},
						val:        "spacebar",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 84, col: 65, offset: 1880},
						val:        "insert",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 84, col: 77, offset: 1892},
						val:        "home",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 85, col: 11, offset: 1910},
						val:        "end",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 85, col: 20, offset: 1919},
						val:        "pageup",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 85, col: 32, offset: 1931},
						val:        "pagedown",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 85, col: 46, offset: 1945},
						val:        "leftalt",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 85, col: 59, offset: 1958},
						val:        "leftctrl",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 85, col: 73, offset: 1972},
						val:        "leftshift",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 86, col: 11, offset: 1995},
						val:        "rightalt",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 86, col: 25, offset: 2009},
						val:        "rightctrl",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 86, col: 40, offset: 2024},
						val:        "rightshift",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 86, col: 56, offset: 2040},
						val:        "leftsuper",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 86, col: 71, offset: 2055},
						val:        "rightsuper",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 87, col: 11, offset: 2079},
						val:        "left",
						ignoreCase: true,
					},
					&litMatcher{
						pos:        position{line: 87, col: 21, offset: 2089},
						val:        "right",
						ignoreCase: true,
					},
				},
			},
		},
		{
			name: "NonZeroDigit",
			pos:  position{line: 89, col: 1, offset: 2099},
			expr: &charClassMatcher{
				pos:        position{line: 89, col: 16, offset: 2114},
				val:        "[1-9]",
				ranges:     []rune{'1', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 90, col: 1, offset: 2120},
			expr: &charClassMatcher{
				pos:        position{line: 90, col: 9, offset: 2128},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "TimeUnit",
			pos:  position{line: 91, col: 1, offset: 2134},
			expr: &choiceExpr{
				pos: position{line: 91, col: 13, offset: 2146},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 91, col: 13, offset: 2146},
						val:        "ns",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 91, col: 20, offset: 2153},
						val:        "us",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 91, col: 27, offset: 2160},
						val:        "µs",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 91, col: 34, offset: 2168},
						val:        "ms",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 91, col: 41, offset: 2175},
						val:        "s",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 91, col: 47, offset: 2181},
						val:        "m",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 91, col: 53, offset: 2187},
						val:        "h",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name:        "_",
			displayName: "\"whitespace\"",
			pos:         position{line: 93, col: 1, offset: 2193},
			expr: &zeroOrMoreExpr{
				pos: position{line: 93, col: 19, offset: 2211},
				expr: &charClassMatcher{
					pos:        position{line: 93, col: 19, offset: 2211},
					val:        "[ \\n\\t\\r]",
					chars:      []rune{' ', '\n', '\t', '\r'},
					ignoreCase: false,
					inverted:   false,
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 95, col: 1, offset: 2223},
			expr: &notExpr{
				pos: position{line: 95, col: 8, offset: 2230},
				expr: &anyMatcher{
					line: 95, col: 9, offset: 2231,
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
		//return fmt.Sprintf("S(%s)", s), nil
		return &specialExpression{string(s.([]byte)), KeyPress}, nil
	}
	return &specialExpression{string(s.([]byte)), t.(KeyAction)}, nil
	//return fmt.Sprintf("S%s(%s)", t, s), nil
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

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errNoMatch is returned if no match could be found.
	errNoMatch = errors.New("no match found")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

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

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
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
}

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

type seqExpr struct {
	pos   position
	exprs []interface{}
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
	pos        position
	val        string
	chars      []rune
	ranges     []rune
	classes    []*unicode.RangeTable
	ignoreCase bool
	inverted   bool
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
	Inner  error
	pos    position
	prefix string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
	}
	p.setOptions(opts)
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

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	recover bool
	debug   bool
	depth   int

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

	// stats
	exprCnt int
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
	p.addErrAt(err, p.pt.position)
}

func (p *parser) addErrAt(err error, pos position) {
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
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String()}
	p.errs.add(pe)
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

	if rn == utf8.RuneError {
		if n == 1 {
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

	// start rule is rule [0]
	p.read() // advance to first rune
	val, ok := p.parseRule(g.rules[0])
	if !ok {
		if len(*p.errs) == 0 {
			// make sure this doesn't go out silently
			p.addErr(errNoMatch)
		}
		return nil, p.errs.err()
	}
	return val, p.errs.err()
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
	var ok bool

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.exprCnt++
	var val interface{}
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
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
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
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position)
		}
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

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn != utf8.RuneError {
		start := p.pt
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	// can't match EOF
	if cur == utf8.RuneError {
		return nil, false
	}
	start := p.pt
	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for _, alt := range ch.alternatives {
		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			return val, ok
		}
	}
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

	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(not.expr)
	p.popV()
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

	var vals []interface{}

	pt := p.pt
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
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

func rangeTable(class string) *unicode.RangeTable {
	if rt, ok := unicode.Categories[class]; ok {
		return rt
	}
	if rt, ok := unicode.Properties[class]; ok {
		return rt
	}
	if rt, ok := unicode.Scripts[class]; ok {
		return rt
	}

	// cannot happen
	panic(fmt.Sprintf("invalid Unicode class: %s", class))
}
