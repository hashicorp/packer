package xmlpath

import (
	"golang.org/x/net/html"
	"encoding/xml"
	"io"
	"strings"
)

// Node is an item in an xml tree that was compiled to
// be processed via xml paths. A node may represent:
//
//     - An element in the xml document (<body>)
//     - An attribute of an element in the xml document (href="...")
//     - A comment in the xml document (<!--...-->)
//     - A processing instruction in the xml document (<?...?>)
//     - Some text within the xml document
//
type Node struct {
	kind nodeKind
	name xml.Name
	attr string
	text []byte

	nodes []Node
	pos   int
	end   int

	up   *Node
	down []*Node
}

type nodeKind int

const (
	anyNode nodeKind = iota
	startNode
	endNode
	attrNode
	textNode
	commentNode
	procInstNode
)

// String returns the string value of node.
//
// The string value of a node is:
//
//     - For element nodes, the concatenation of all text nodes within the element.
//     - For text nodes, the text itself.
//     - For attribute nodes, the attribute value.
//     - For comment nodes, the text within the comment delimiters.
//     - For processing instruction nodes, the content of the instruction.
//
func (node *Node) String() string {
	if node.kind == attrNode {
		return node.attr
	}
	return string(node.Bytes())
}

// Bytes returns the string value of node as a byte slice.
// See Node.String for a description of what the string value of a node is.
func (node *Node) Bytes() []byte {
	if node.kind == attrNode {
		return []byte(node.attr)
	}
	if node.kind != startNode {
		return node.text
	}
	size := 0
	for i := node.pos; i < node.end; i++ {
		if node.nodes[i].kind == textNode {
			size += len(node.nodes[i].text)
		}
	}
	text := make([]byte, 0, size)
	for i := node.pos; i < node.end; i++ {
		if node.nodes[i].kind == textNode {
			text = append(text, node.nodes[i].text...)
		}
	}
	return text
}

// equals returns whether the string value of node is equal to s,
// without allocating memory.
func (node *Node) equals(s string) bool {
	if node.kind == attrNode {
		return s == node.attr
	}
	if node.kind != startNode {
		if len(s) != len(node.text) {
			return false
		}
		for i := range s {
			if s[i] != node.text[i] {
				return false
			}
		}
		return true
	}
	si := 0
	for i := node.pos; i < node.end; i++ {
		if node.nodes[i].kind == textNode {
			for _, c := range node.nodes[i].text {
				if si > len(s) {
					return false
				}
				if s[si] != c {
					return false
				}
				si++
			}
		}
	}
	return si == len(s)
}

// contains returns whether the string value of node contains s,
// without allocating memory.
func (node *Node) contains(s string) (ok bool) {
	if len(s) == 0 {
		return true
	}
	if node.kind == attrNode {
		return strings.Contains(node.attr, s)
	}
	s0 := s[0]
	for i := node.pos; i < node.end; i++ {
		if node.nodes[i].kind == textNode {
			text := node.nodes[i].text
		NextTry:
			for ci, c := range text {
				if c != s0 {
					continue
				}
				si := 1
				for ci++; ci < len(text) && si < len(s); ci++ {
					if s[si] != text[ci] {
						continue NextTry
					}
					si++
				}
				if si == len(s) {
					return true
				}
				for j := i + 1; j < node.end; j++ {
					if node.nodes[j].kind == textNode {
						for _, c := range node.nodes[j].text {
							if s[si] != c {
								continue NextTry
							}
							si++
							if si == len(s) {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

// Parse reads an xml document from r, parses it, and returns its root node.
func Parse(r io.Reader) (*Node, error) {
	return ParseDecoder(xml.NewDecoder(r))
}

// ParseDecoder parses the xml document being decoded by d and returns
// its root node.
func ParseDecoder(d *xml.Decoder) (*Node, error) {
	var nodes []Node
	var text []byte

	// The root node.
	nodes = append(nodes, Node{kind: startNode})

	for {
		t, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := t.(type) {
		case xml.EndElement:
			nodes = append(nodes, Node{
				kind: endNode,
			})
		case xml.StartElement:
			nodes = append(nodes, Node{
				kind: startNode,
				name: t.Name,
			})
			for _, attr := range t.Attr {
				nodes = append(nodes, Node{
					kind: attrNode,
					name: attr.Name,
					attr: attr.Value,
				})
			}
		case xml.CharData:
			texti := len(text)
			text = append(text, t...)
			nodes = append(nodes, Node{
				kind: textNode,
				text: text[texti : texti+len(t)],
			})
		case xml.Comment:
			texti := len(text)
			text = append(text, t...)
			nodes = append(nodes, Node{
				kind: commentNode,
				text: text[texti : texti+len(t)],
			})
		case xml.ProcInst:
			texti := len(text)
			text = append(text, t.Inst...)
			nodes = append(nodes, Node{
				kind: procInstNode,
				name: xml.Name{Local: t.Target},
				text: text[texti : texti+len(t.Inst)],
			})
		}
	}

	// Close the root node.
	nodes = append(nodes, Node{kind: endNode})

	stack := make([]*Node, 0, len(nodes))
	downs := make([]*Node, len(nodes))
	downCount := 0

	for pos := range nodes {

		switch nodes[pos].kind {

		case startNode, attrNode, textNode, commentNode, procInstNode:
			node := &nodes[pos]
			node.nodes = nodes
			node.pos = pos
			if len(stack) > 0 {
				node.up = stack[len(stack)-1]
			}
			if node.kind == startNode {
				stack = append(stack, node)
			} else {
				node.end = pos + 1
			}

		case endNode:
			node := stack[len(stack)-1]
			node.end = pos
			stack = stack[:len(stack)-1]

			// Compute downs. Doing that here is what enables the
			// use of a slice of a contiguous pre-allocated block.
			node.down = downs[downCount:downCount]
			for i := node.pos + 1; i < node.end; i++ {
				if nodes[i].up == node {
					switch nodes[i].kind {
					case startNode, textNode, commentNode, procInstNode:
						node.down = append(node.down, &nodes[i])
						downCount++
					}
				}
			}
			if len(stack) == 0 {
				return node, nil
			}
		}
	}
	return nil, io.EOF
}

// ParseHTML reads an HTML document from r, parses it using a proper HTML
// parser, and returns its root node.
//
// The document will be processed as a properly structured HTML document,
// emulating the behavior of a browser when processing it. This includes
// putting the content inside proper <html> and <body> tags, if the
// provided text misses them.
func ParseHTML(r io.Reader) (*Node, error) {
	ns, err := html.ParseFragment(r, nil)
	if err != nil {
		return nil, err
	}

	var nodes []Node
	var text []byte

	n := ns[0]

	// The root node.
	nodes = append(nodes, Node{kind: startNode})

	for n != nil {
		switch n.Type {
		case html.DocumentNode:
		case html.ElementNode:
			nodes = append(nodes, Node{
				kind: startNode,
				name: xml.Name{Local: n.Data, Space: n.Namespace},
			})
			for _, attr := range n.Attr {
				nodes = append(nodes, Node{
					kind: attrNode,
					name: xml.Name{Local: attr.Key, Space: attr.Namespace},
					attr: attr.Val,
				})
			}
		case html.TextNode:
			texti := len(text)
			text = append(text, n.Data...)
			nodes = append(nodes, Node{
				kind: textNode,
				text: text[texti : texti+len(n.Data)],
			})
		case html.CommentNode:
			texti := len(text)
			text = append(text, n.Data...)
			nodes = append(nodes, Node{
				kind: commentNode,
				text: text[texti : texti+len(n.Data)],
			})
		}

		if n.FirstChild != nil {
			n = n.FirstChild
			continue
		}

		for n != nil {
			if n.Type == html.ElementNode {
				nodes = append(nodes, Node{kind: endNode})
			}
			if n.NextSibling != nil {
				n = n.NextSibling
				break
			}
			n = n.Parent
		}
	}

	// Close the root node.
	nodes = append(nodes, Node{kind: endNode})

	stack := make([]*Node, 0, len(nodes))
	downs := make([]*Node, len(nodes))
	downCount := 0

	for pos := range nodes {

		switch nodes[pos].kind {

		case startNode, attrNode, textNode, commentNode, procInstNode:
			node := &nodes[pos]
			node.nodes = nodes
			node.pos = pos
			if len(stack) > 0 {
				node.up = stack[len(stack)-1]
			}
			if node.kind == startNode {
				stack = append(stack, node)
			} else {
				node.end = pos + 1
			}

		case endNode:
			node := stack[len(stack)-1]
			node.end = pos
			stack = stack[:len(stack)-1]

			// Compute downs. Doing that here is what enables the
			// use of a slice of a contiguous pre-allocated block.
			node.down = downs[downCount:downCount]
			for i := node.pos + 1; i < node.end; i++ {
				if nodes[i].up == node {
					switch nodes[i].kind {
					case startNode, textNode, commentNode, procInstNode:
						node.down = append(node.down, &nodes[i])
						downCount++
					}
				}
			}
			if len(stack) == 0 {
				return node, nil
			}
		}
	}
	return nil, io.EOF
}
