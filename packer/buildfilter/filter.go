// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

// Package buildfilter implements the parser and matcher for the
// `packer build -filter=...` flag. Filters select CoreBuilds by their
// declared tags and labels (key/value metadata).
//
// Grammar:
//
//	<filter>     := <key> <op> <value-list>
//	<key>        := "tags" | <label-key>
//	<op>         := "="    all of   (every listed value must match)
//	              | "~="   any of   (at least one listed value matches)
//	              | "!="   none of  (no listed value matches)
//	<value-list> := <value> ("," <value>)*
//
// Values support glob patterns via github.com/gobwas/glob, matching the
// behavior of the existing -only/-except flags.
//
// Multiple -filter flags are AND-ed together, like Kubernetes label
// selectors: a build must satisfy every filter expression to be selected.
package buildfilter

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
)

// Op is a filter operator.
type Op int

const (
	// OpAll requires every listed value to match (implicit AND within values).
	OpAll Op = iota
	// OpAny requires at least one listed value to match.
	OpAny
	// OpNone requires no listed value to match.
	OpNone
)

func (o Op) String() string {
	switch o {
	case OpAll:
		return "="
	case OpAny:
		return "~="
	case OpNone:
		return "!="
	}
	return "?"
}

// Expr is a parsed filter expression.
type Expr struct {
	// Key is either the literal string "tags" or a label key.
	Key string
	Op  Op
	// Values holds the compiled glob patterns the user supplied.
	Values []glob.Glob
	// Raw preserves the original user-supplied value list for error messages.
	Raw []string
}

// Taggable is the interface builds must satisfy to be filtered. CoreBuild
// implements this.
type Taggable interface {
	FilterTags() []string
	FilterLabels() map[string]string
}

// Parse compiles a slice of raw filter strings (as supplied via -filter on
// the CLI) into a slice of Expr that can be evaluated with Match.
//
// An empty raw slice returns (nil, nil) and matches every build.
func Parse(raw []string) ([]Expr, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]Expr, 0, len(raw))
	for _, s := range raw {
		e, err := parseOne(s)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func parseOne(s string) (Expr, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return Expr{}, fmt.Errorf("empty -filter expression")
	}

	// Longer operator tokens must be tried first so "~=" and "!=" aren't
	// mis-identified as "=".
	var key, op, rest string
	switch {
	case containsOp(trimmed, "~="):
		key, rest = splitOnce(trimmed, "~=")
		op = "~="
	case containsOp(trimmed, "!="):
		key, rest = splitOnce(trimmed, "!=")
		op = "!="
	case containsOp(trimmed, "="):
		key, rest = splitOnce(trimmed, "=")
		op = "="
	default:
		return Expr{}, fmt.Errorf("invalid -filter %q: expected KEY=VALUE, KEY~=VALUE, or KEY!=VALUE", s)
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return Expr{}, fmt.Errorf("invalid -filter %q: empty key", s)
	}
	if !validKey(key) {
		return Expr{}, fmt.Errorf("invalid -filter key %q: must match [A-Za-z_][A-Za-z0-9_.-]*", key)
	}

	rawValues := splitValues(rest)
	if len(rawValues) == 0 {
		return Expr{}, fmt.Errorf("invalid -filter %q: empty value list", s)
	}

	globs := make([]glob.Glob, 0, len(rawValues))
	for _, v := range rawValues {
		g, err := glob.Compile(v)
		if err != nil {
			return Expr{}, fmt.Errorf("invalid -filter %q: bad glob %q: %s", s, v, err)
		}
		globs = append(globs, g)
	}

	e := Expr{Key: key, Values: globs, Raw: rawValues}
	switch op {
	case "=":
		e.Op = OpAll
	case "~=":
		e.Op = OpAny
	case "!=":
		e.Op = OpNone
	}
	return e, nil
}

// containsOp reports whether s contains op as a literal substring. A small
// helper to keep parseOne readable.
func containsOp(s, op string) bool { return strings.Contains(s, op) }

// splitOnce splits s on the first occurrence of sep, returning the two halves.
func splitOnce(s, sep string) (string, string) {
	i := strings.Index(s, sep)
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+len(sep):]
}

// splitValues splits a comma-separated value list, trimming whitespace and
// dropping empty entries.
func splitValues(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func validKey(k string) bool {
	if k == "" {
		return false
	}
	for i, r := range k {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r == '_':
		case i > 0 && (r >= '0' && r <= '9' || r == '.' || r == '-'):
		default:
			return false
		}
	}
	return true
}

// Match reports whether b satisfies every expression in exprs. An empty
// exprs slice matches any build.
func Match(b Taggable, exprs []Expr) bool {
	if len(exprs) == 0 {
		return true
	}
	tags := b.FilterTags()
	labels := b.FilterLabels()
	for _, e := range exprs {
		if !matchOne(e, tags, labels) {
			return false
		}
	}
	return true
}

func matchOne(e Expr, tags []string, labels map[string]string) bool {
	if e.Key == "tags" {
		return matchTags(e, tags)
	}
	return matchLabel(e, labels[e.Key])
}

func matchTags(e Expr, tags []string) bool {
	anyHit := func(g glob.Glob) bool {
		for _, t := range tags {
			if g.Match(t) {
				return true
			}
		}
		return false
	}
	switch e.Op {
	case OpAll:
		for _, g := range e.Values {
			if !anyHit(g) {
				return false
			}
		}
		return true
	case OpAny:
		for _, g := range e.Values {
			if anyHit(g) {
				return true
			}
		}
		return false
	case OpNone:
		for _, g := range e.Values {
			if anyHit(g) {
				return false
			}
		}
		return true
	}
	return false
}

func matchLabel(e Expr, v string) bool {
	// Label equality/inequality: the label value must match the supplied
	// glob(s). A build with no such label fails OpAll/OpAny but satisfies
	// OpNone (there is nothing to exclude it on).
	switch e.Op {
	case OpAll:
		if v == "" {
			return false
		}
		for _, g := range e.Values {
			if !g.Match(v) {
				return false
			}
		}
		return true
	case OpAny:
		if v == "" {
			return false
		}
		for _, g := range e.Values {
			if g.Match(v) {
				return true
			}
		}
		return false
	case OpNone:
		if v == "" {
			return true
		}
		for _, g := range e.Values {
			if g.Match(v) {
				return false
			}
		}
		return true
	}
	return false
}
