package xmlpath

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

// Path is a compiled path that can be applied to a context
// node to obtain a matching node set.
// A single Path can be applied concurrently to any number
// of context nodes.
type Path struct {
	path  string
	steps []pathStep
}

// Iter returns an iterator that goes over the list of nodes
// that p matches on the given context.
func (p *Path) Iter(context *Node) *Iter {
	iter := Iter{
		make([]pathStepState, len(p.steps)),
		make([]bool, len(context.nodes)),
	}
	for i := range p.steps {
		iter.state[i].step = &p.steps[i]
	}
	iter.state[0].init(context)
	return &iter
}

// Exists returns whether any nodes match p on the given context.
func (p *Path) Exists(context *Node) bool {
	return p.Iter(context).Next()
}

// String returns the string value of the first node matched
// by p on the given context.
//
// See the documentation of Node.String.
func (p *Path) String(context *Node) (s string, ok bool) {
	iter := p.Iter(context)
	if iter.Next() {
		return iter.Node().String(), true
	}
	return "", false
}

// Bytes returns as a byte slice the string value of the first
// node matched by p on the given context.
//
// See the documentation of Node.String.
func (p *Path) Bytes(node *Node) (b []byte, ok bool) {
	iter := p.Iter(node)
	if iter.Next() {
		return iter.Node().Bytes(), true
	}
	return nil, false
}

// Iter iterates over node sets.
type Iter struct {
	state []pathStepState
	seen  []bool
}

// Node returns the current node.
// Must only be called after Iter.Next returns true.
func (iter *Iter) Node() *Node {
	state := iter.state[len(iter.state)-1]
	if state.pos == 0 {
		panic("Iter.Node called before Iter.Next")
	}
	if state.node == nil {
		panic("Iter.Node called after Iter.Next false")
	}
	return state.node
}

// Next iterates to the next node in the set, if any, and
// returns whether there is a node available.
func (iter *Iter) Next() bool {
	tip := len(iter.state) - 1
outer:
	for {
		for !iter.state[tip].next() {
			tip--
			if tip == -1 {
				return false
			}
		}
		for tip < len(iter.state)-1 {
			tip++
			iter.state[tip].init(iter.state[tip-1].node)
			if !iter.state[tip].next() {
				tip--
				continue outer
			}
		}
		if iter.seen[iter.state[tip].node.pos] {
			continue
		}
		iter.seen[iter.state[tip].node.pos] = true
		return true
	}
	panic("unreachable")
}

type pathStepState struct {
	step *pathStep
	node *Node
	pos  int
	idx  int
	aux  int
}

func (s *pathStepState) init(node *Node) {
	s.node = node
	s.pos = 0
	s.idx = 0
	s.aux = 0
}

func (s *pathStepState) next() bool {
	for s._next() {
		s.pos++
		if s.step.pred == nil || s.test(s.step.pred) {
			return true
		}
	}
	return false
}

func (s *pathStepState) test(pred predicate) bool {
	switch pred := pred.(type) {
	case positionPredicate:
		if pred.pos == s.pos {
			return true
		}
	case existsPredicate:
		if pred.path.Exists(s.node) {
			return true
		}
	case equalsPredicate:
		iter := pred.path.Iter(s.node)
		for iter.Next() {
			if iter.Node().equals(pred.value) {
				return true
			}
		}
	case containsPredicate:
		iter := pred.path.Iter(s.node)
		for iter.Next() {
			if iter.Node().contains(pred.value) {
				return true
			}
		}
	case notPredicate:
		iter := pred.path.Iter(s.node)
		if !iter.Next() {
			return true
		}
	case andPredicate:
		for _, sub := range pred.sub {
			if !s.test(sub) {
				return false
			}
		}
		return true
	case orPredicate:
		for _, sub := range pred.sub {
			if s.test(sub) {
				return true
			}
		}
	default:
		panic(fmt.Sprintf("internal error: unknown predicate type: %#v", s.step.pred))
	}
	return false
}

func (s *pathStepState) _next() bool {
	if s.node == nil {
		return false
	}
	if s.step.root && s.idx == 0 {
		for s.node.up != nil {
			s.node = s.node.up
		}
	}

	switch s.step.axis {

	case "self":
		if s.idx == 0 && s.step.match(s.node) {
			s.idx++
			return true
		}

	case "parent":
		if s.idx == 0 && s.node.up != nil && s.step.match(s.node.up) {
			s.idx++
			s.node = s.node.up
			return true
		}

	case "ancestor", "ancestor-or-self":
		if s.idx == 0 && s.step.axis == "ancestor-or-self" {
			s.idx++
			if s.step.match(s.node) {
				return true
			}
		}
		for s.node.up != nil {
			s.node = s.node.up
			s.idx++
			if s.step.match(s.node) {
				return true
			}
		}

	case "child":
		var down []*Node
		if s.idx == 0 {
			down = s.node.down
		} else {
			down = s.node.up.down
		}
		for s.idx < len(down) {
			node := down[s.idx]
			s.idx++
			if s.step.match(node) {
				s.node = node
				return true
			}
		}

	case "descendant", "descendant-or-self":
		if s.idx == 0 {
			s.idx = s.node.pos
			s.aux = s.node.end
			if s.step.axis == "descendant" {
				s.idx++
			}
		}
		for s.idx < s.aux {
			node := &s.node.nodes[s.idx]
			s.idx++
			if node.kind == attrNode {
				continue
			}
			if s.step.match(node) {
				s.node = node
				return true
			}
		}

	case "following":
		if s.idx == 0 {
			s.idx = s.node.end
		}
		for s.idx < len(s.node.nodes) {
			node := &s.node.nodes[s.idx]
			s.idx++
			if node.kind == attrNode {
				continue
			}
			if s.step.match(node) {
				s.node = node
				return true
			}
		}

	case "following-sibling":
		var down []*Node
		if s.node.up != nil {
			down = s.node.up.down
			if s.idx == 0 {
				for s.idx < len(down) {
					node := down[s.idx]
					s.idx++
					if node == s.node {
						break
					}
				}
			}
		}
		for s.idx < len(down) {
			node := down[s.idx]
			s.idx++
			if s.step.match(node) {
				s.node = node
				return true
			}
		}

	case "preceding":
		if s.idx == 0 {
			s.aux = s.node.pos // Detect ancestors.
			s.idx = s.node.pos - 1
		}
		for s.idx >= 0 {
			node := &s.node.nodes[s.idx]
			s.idx--
			if node.kind == attrNode {
				continue
			}
			if node == s.node.nodes[s.aux].up {
				s.aux = s.node.nodes[s.aux].up.pos
				continue
			}
			if s.step.match(node) {
				s.node = node
				return true
			}
		}

	case "preceding-sibling":
		var down []*Node
		if s.node.up != nil {
			down = s.node.up.down
			if s.aux == 0 {
				s.aux = 1
				for s.idx < len(down) {
					node := down[s.idx]
					s.idx++
					if node == s.node {
						s.idx--
						break
					}
				}
			}
		}
		for s.idx >= 0 {
			node := down[s.idx]
			s.idx--
			if s.step.match(node) {
				s.node = node
				return true
			}
		}

	case "attribute":
		if s.idx == 0 {
			s.idx = s.node.pos + 1
			s.aux = s.node.end
		}
		for s.idx < s.aux {
			node := &s.node.nodes[s.idx]
			s.idx++
			if node.kind != attrNode {
				break
			}
			if s.step.match(node) {
				s.node = node
				return true
			}
		}

	}

	s.node = nil
	return false
}

type positionPredicate struct {
	pos int
}

type existsPredicate struct {
	path *Path
}

type equalsPredicate struct {
	path  *Path
	value string
}

type containsPredicate struct {
	path  *Path
	value string
}

type notPredicate struct {
	path *Path
}

type andPredicate struct {
	sub []predicate
}

type orPredicate struct {
	sub []predicate
}

// predicate is a marker interface for predicate types.
type predicate interface {
	predicate()
}

func (positionPredicate) predicate() {}
func (existsPredicate) predicate()   {}
func (equalsPredicate) predicate()   {}
func (containsPredicate) predicate() {}
func (notPredicate) predicate()      {}
func (andPredicate) predicate()      {}
func (orPredicate) predicate()       {}

type pathStep struct {
	root bool
	axis string
	name string
	kind nodeKind
	pred predicate
}

func (step *pathStep) match(node *Node) bool {
	return node.kind != endNode &&
		(step.kind == anyNode || step.kind == node.kind) &&
		(step.name == "*" || node.name.Local == step.name)
}

// MustCompile returns the compiled path, and panics if
// there are any errors.
func MustCompile(path string) *Path {
	e, err := Compile(path)
	if err != nil {
		panic(err)
	}
	return e
}

// Compile returns the compiled path.
func Compile(path string) (*Path, error) {
	c := pathCompiler{path, 0}
	if path == "" {
		return nil, c.errorf("empty path")
	}
	p, err := c.parsePath()
	if err != nil {
		return nil, err
	}
	return p, nil
}

type pathCompiler struct {
	path string
	i    int
}

func (c *pathCompiler) errorf(format string, args ...interface{}) error {
	return fmt.Errorf("compiling xml path %q:%d: %s", c.path, c.i, fmt.Sprintf(format, args...))
}

func (c *pathCompiler) parsePath() (path *Path, err error) {
	var steps []pathStep
	var start = c.i
	for {
		step := pathStep{axis: "child"}

		c.skipSpaces()
		if c.i == 0 && c.skipByte('/') {
			c.skipSpaces()
			step.root = true
			if len(c.path) == 1 {
				step.name = "*"
			}
		}
		if c.peekByte('/') {
			step.axis = "descendant-or-self"
			step.name = "*"
		} else if c.skipByte('@') {
			mark := c.i
			if !c.skipName() {
				return nil, c.errorf("missing name after @")
			}
			step.axis = "attribute"
			step.name = c.path[mark:c.i]
			step.kind = attrNode
		} else {
			mark := c.i
			if c.skipName() {
				step.name = c.path[mark:c.i]
				c.skipSpaces()
			}
			if step.name == "" {
				return nil, c.errorf("missing name")
			} else if step.name == "*" {
				step.kind = startNode
			} else if step.name == "." {
				step.axis = "self"
				step.name = "*"
			} else if step.name == ".." {
				step.axis = "parent"
				step.name = "*"
			} else {
				if c.skipByte(':') {
					if !c.skipByte(':') {
						return nil, c.errorf("missing ':'")
					}
					c.skipSpaces()
					switch step.name {
					case "attribute":
						step.kind = attrNode
					case "self", "child", "parent":
					case "descendant", "descendant-or-self":
					case "ancestor", "ancestor-or-self":
					case "following", "following-sibling":
					case "preceding", "preceding-sibling":
					default:
						return nil, c.errorf("unsupported axis: %q", step.name)
					}
					step.axis = step.name

					mark = c.i
					if !c.skipName() {
						return nil, c.errorf("missing name")
					}
					step.name = c.path[mark:c.i]

					c.skipSpaces()
				}
				if c.skipByte('(') {
					c.skipSpaces()
					conflict := step.kind != anyNode
					switch step.name {
					case "node":
						// must be anyNode
					case "text":
						step.kind = textNode
					case "comment":
						step.kind = commentNode
					case "processing-instruction":
						step.kind = procInstNode
					default:
						return nil, c.errorf("unsupported expression: %s()", step.name)
					}
					if conflict {
						return nil, c.errorf("%s() cannot succeed on axis %q", step.name, step.axis)
					}

					name := step.name
					literal, err := c.parseLiteral()
					if err == errNoLiteral {
						step.name = "*"
					} else if err != nil {
						return nil, c.errorf("%v", err)
					} else if step.kind == procInstNode {
						c.skipSpaces()
						step.name = literal
					} else {
						return nil, c.errorf("%s() has no arguments", name)
					}
					if !c.skipByte(')') {
						return nil, c.errorf("%s() missing ')'", name)
					}
					c.skipSpaces()
				} else if step.name == "*" && step.kind == anyNode {
					step.kind = startNode
				}
			}
		}
		if c.skipByte('[') {
			c.skipSpaces()
			type state struct {
				sub []predicate
				and bool
			}
			var stack []state
			var sub []predicate
			var and bool
		NextPred:
			if c.skipByte('(') {
				stack = append(stack, state{sub, and})
				sub = nil
				and = false
			}
			var next predicate
			if pos, ok := c.parseInt(); ok {
				if pos == 0 {
					return nil, c.errorf("positions start at 1")
				}
				next = positionPredicate{pos}
			} else if c.skipString("contains(") {
				path, err := c.parsePath()
				if err != nil {
					return nil, err
				}
				c.skipSpaces()
				if !c.skipByte(',') {
					return nil, c.errorf("contains() expected ',' followed by a literal string")
				}
				c.skipSpaces()
				value, err := c.parseLiteral()
				if err != nil {
					return nil, err
				}
				c.skipSpaces()
				if !c.skipByte(')') {
					return nil, c.errorf("contains() missing ')'")
				}
				next = containsPredicate{path, value}
			} else if c.skipString("not(") {
				// TODO Generalize to handle any predicate expression.
				path, err := c.parsePath()
				if err != nil {
					return nil, err
				}
				c.skipSpaces()
				if !c.skipByte(')') {
					return nil, c.errorf("not() missing ')'")
				}
				next = notPredicate{path}
			} else {
				path, err := c.parsePath()
				if err != nil {
					return nil, err
				}
				if path.path[0] == '-' {
					if _, err = strconv.Atoi(path.path); err == nil {
						return nil, c.errorf("positions must be positive")
					}
				}
				c.skipSpaces()
				if c.skipByte('=') {
					c.skipSpaces()
					value, err := c.parseLiteral()
					if err != nil {
						return nil, c.errorf("%v", err)
					}
					next = equalsPredicate{path, value}
				} else {
					next = existsPredicate{path}
				}
			}
		HandleNext:
			if and {
				p := sub[len(sub)-1].(andPredicate)
				p.sub = append(p.sub, next)
				sub[len(sub)-1] = p
			} else {
				sub = append(sub, next)
			}
			if c.skipSpaces() {
				mark := c.i
				if c.skipString("and") && c.skipSpaces() {
					if !and {
						and = true
						sub[len(sub)-1] = andPredicate{[]predicate{sub[len(sub)-1]}}
					}
					goto NextPred
				} else if c.skipString("or") && c.skipSpaces() {
					and = false
					goto NextPred
				} else {
					c.i = mark
				}
			}
			if c.skipByte(')') {
				if len(stack) == 0 {
					return nil, c.errorf("unexpected ')'")
				}
				if len(sub) == 1 {
					next = sub[0]
				} else {
					next = orPredicate{sub}
				}
				s := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				sub = s.sub
				and = s.and
				goto HandleNext
			}
			if len(stack) > 0 {
				return nil, c.errorf("expected ')'")
			}
			if len(sub) == 1 {
				step.pred = sub[0]
			} else {
				step.pred = orPredicate{sub}
			}
			if !c.skipByte(']') {
				return nil, c.errorf("expected ']'")
			}
			c.skipSpaces()
		}
		steps = append(steps, step)
		//fmt.Printf("step: %#v\n", step)
		if !c.skipByte('/') {
			if (start == 0 || start == c.i) && c.i < len(c.path) {
				return nil, c.errorf("unexpected %q", c.path[c.i])
			}
			return &Path{steps: steps, path: c.path[start:c.i]}, nil
		}
	}
	panic("unreachable")
}

var errNoLiteral = fmt.Errorf("expected a literal string")

func (c *pathCompiler) parseLiteral() (string, error) {
	if c.skipByte('"') {
		mark := c.i
		if !c.skipByteFind('"') {
			return "", fmt.Errorf(`missing '"'`)
		}
		return c.path[mark : c.i-1], nil
	}
	if c.skipByte('\'') {
		mark := c.i
		if !c.skipByteFind('\'') {
			return "", fmt.Errorf(`missing "'"`)
		}
		return c.path[mark : c.i-1], nil
	}
	return "", errNoLiteral
}

func (c *pathCompiler) parseInt() (v int, ok bool) {
	mark := c.i
	for c.i < len(c.path) && c.path[c.i] >= '0' && c.path[c.i] <= '9' {
		v *= 10
		v += int(c.path[c.i]) - '0'
		c.i++
	}
	if c.i == mark {
		return 0, false
	}
	return v, true
}

func (c *pathCompiler) skipByte(b byte) bool {
	if c.i < len(c.path) && c.path[c.i] == b {
		c.i++
		return true
	}
	return false
}

func (c *pathCompiler) skipByteFind(b byte) bool {
	for i := c.i; i < len(c.path); i++ {
		if c.path[i] == b {
			c.i = i + 1
			return true
		}
	}
	return false
}

func (c *pathCompiler) peekByte(b byte) bool {
	return c.i < len(c.path) && c.path[c.i] == b
}

func (c *pathCompiler) skipSpaces() bool {
	mark := c.i
	for c.i < len(c.path) {
		if c.path[c.i] != ' ' {
			break
		}
		c.i++
	}
	return c.i != mark
}

func (c *pathCompiler) skipString(s string) bool {
	if c.i+len(s) <= len(c.path) && c.path[c.i:c.i+len(s)] == s {
		c.i += len(s)
		return true
	}
	return false
}

func (c *pathCompiler) skipName() bool {
	if c.i >= len(c.path) {
		return false
	}
	if c.path[c.i] == '*' {
		c.i++
		return true
	}
	start := c.i
	for c.i < len(c.path) && (c.path[c.i] >= utf8.RuneSelf || isNameByte(c.path[c.i])) {
		c.i++
	}
	return c.i > start
}

func isNameByte(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '_' || c == '.' || c == '-'
}
