// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package dag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

// AcyclicGraph is a specialization of Graph that cannot have cycles.
type AcyclicGraph struct {
	Graph
}

// WalkFunc is the callback used for walking the graph.
type WalkFunc func(Vertex) hcl.Diagnostics

// DepthWalkFunc is a walk function that also receives the current depth of the
// walk as an argument
type DepthWalkFunc func(Vertex, int) error

func (g *AcyclicGraph) DirectedGraph() Grapher {
	return g
}

// Validate validates the DAG. A DAG is valid if it has no cycles or self-referencing vertex.
func (g *AcyclicGraph) Validate() error {
	// Look for cycles of more than 1 component
	var err error
	cycles := g.Cycles()
	if len(cycles) > 0 {
		for _, cycle := range cycles {
			cycleStr := make([]string, len(cycle))
			for j, vertex := range cycle {
				cycleStr[j] = VertexName(vertex)
			}

			err = errors.Join(err, fmt.Errorf(
				"Cycle: %s", strings.Join(cycleStr, ", ")))
		}
	}

	// Look for cycles to self
	for _, e := range g.Edges() {
		if e.Source() == e.Target() {
			err = errors.Join(err, fmt.Errorf(
				"Self reference: %s", VertexName(e.Source())))
		}
	}

	return err
}

// Cycles reports any cycles between graph nodes.
// Self-referencing nodes are not reported, and must be detected separately.
func (g *AcyclicGraph) Cycles() [][]Vertex {
	var cycles [][]Vertex
	for _, cycle := range StronglyConnected(&g.Graph) {
		if len(cycle) > 1 {
			cycles = append(cycles, cycle)
		}
	}
	return cycles
}

type walkType uint64

const (
	depthFirst walkType = 1 << iota
	breadthFirst
	downOrder
	upOrder
)

// ReverseTopologicalOrder returns a topological sort of the given graph, with
// target vertices ordered before the sources of their edges. The nodes are not
// sorted, and any valid order may be returned. This function will panic if it
// encounters a cycle.
func (g *AcyclicGraph) ReverseTopologicalOrder() []Vertex {
	return g.topoOrder(downOrder)
}

func (g *AcyclicGraph) topoOrder(order walkType) []Vertex {
	// Use a dfs-based sorting algorithm, similar to that used in
	// TransitiveReduction.
	sorted := make([]Vertex, 0, len(g.vertices))

	// tmp track the current working node to check for cycles
	tmp := map[Vertex]bool{}

	// perm tracks completed nodes to end the recursion
	perm := map[Vertex]bool{}

	var visit func(v Vertex)

	visit = func(v Vertex) {
		if perm[v] {
			return
		}

		if tmp[v] {
			panic("cycle found in dag")
		}

		tmp[v] = true
		var next Set
		switch {
		case order&downOrder != 0:
			next = g.downEdgesNoCopy(v)
		case order&upOrder != 0:
			next = g.upEdgesNoCopy(v)
		default:
			panic(fmt.Sprintln("invalid order", order))
		}

		for _, u := range next {
			visit(u)
		}

		tmp[v] = false
		perm[v] = true
		sorted = append(sorted, v)
	}

	for _, v := range g.Vertices() {
		visit(v)
	}

	return sorted
}
