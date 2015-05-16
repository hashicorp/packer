package interpolate

import (
	"fmt"
	"text/template"
	"text/template/parse"
)

// functionsCalled returns a map (to be used as a set) of the functions
// that are called from the given text template.
func functionsCalled(t *template.Template) map[string]struct{} {
	result := make(map[string]struct{})
	functionsCalledWalk(t.Tree.Root, result)
	return result
}

func functionsCalledWalk(raw parse.Node, r map[string]struct{}) {
	switch node := raw.(type) {
	case *parse.ActionNode:
		functionsCalledWalk(node.Pipe, r)
	case *parse.CommandNode:
		if in, ok := node.Args[0].(*parse.IdentifierNode); ok {
			r[in.Ident] = struct{}{}
		}

		for _, n := range node.Args[1:] {
			functionsCalledWalk(n, r)
		}
	case *parse.ListNode:
		for _, n := range node.Nodes {
			functionsCalledWalk(n, r)
		}
	case *parse.PipeNode:
		for _, n := range node.Cmds {
			functionsCalledWalk(n, r)
		}
	case *parse.StringNode, *parse.TextNode:
		// Ignore
	default:
		panic(fmt.Sprintf("unknown type: %T", node))
	}
}
