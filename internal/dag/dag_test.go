// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package dag

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestAcyclicGraphValidate(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(3, 2))
	g.Connect(BasicEdge(3, 1))

	if err := g.Validate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestAcyclicGraphValidate_cycle(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(3, 2))
	g.Connect(BasicEdge(3, 1))
	g.Connect(BasicEdge(1, 2))
	g.Connect(BasicEdge(2, 1))

	if err := g.Validate(); err == nil {
		t.Fatal("should error")
	}
}

func TestAcyclicGraphValidate_cycleSelf(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Connect(BasicEdge(1, 1))

	if err := g.Validate(); err == nil {
		t.Fatal("should error")
	}
}
