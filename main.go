package main

import (
	"context"
	"fmt"
	"tuscany/pkg/graph"
)

type data struct {
	firstComputation  int
	secondComputation int
	thirdComputation  int
	finalComputation  int
}

func main() {
	d := &data{}

	scheduler := graph.NewScheduler(nil)

	g := graph.NewDAG(scheduler)

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
	g.AddEdges(second, first)

	third := graph.NewNode("computing_second_sum", func(ctx context.Context) error {
		d.thirdComputation = d.firstComputation + 2
		return nil
	})
	g.AddNode(third)
	g.AddEdges(third, first)

	final := graph.NewNode("computing_final_sum", func(ctx context.Context) error {
		d.finalComputation = d.secondComputation + d.thirdComputation + 1
		return nil
	})
	g.AddNode(final)
	g.AddEdges(final, second, third)

	ctx := context.Background()

	err := g.Compute(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("TADA: %d.\n", d.finalComputation)
}
