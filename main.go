package main

import (
	"context"
	"fmt"
	"tuscany/pkg/graph"
)

type data struct {
	firstComputation  int
	secondComputation int
	finalComputation  int
}

func main() {
	d := &data{}

	g := graph.NewDAG()

	first := graph.NewNode("computing_first_sum", func(ctx context.Context) error {
		d.firstComputation = 1 + 1
		return nil
	})
	g.AddNode(first)

	second := graph.NewNode("computing_second_sum", func(ctx context.Context) error {
		d.secondComputation = d.firstComputation + 1
		return nil
	})
	g.AddNode(second)
	g.AddEdges(first, second)

	final := graph.NewNode("computing_final_sum", func(ctx context.Context) error {
		d.finalComputation = d.secondComputation + 1
		return nil
	})
	g.AddNode(final)
	g.AddEdges(second, final)

	ctx := context.Background()

	err := g.Compute(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("TADA: %d.", d.finalComputation)
}
