package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDAG(t *testing.T) {
	t.Run("diamant_dag_success", func(t *testing.T) {
		var (
			firstComputation  int
			secondComputation int
			thirdComputation  int
			finalComputation  int
		)

		scheduler := NewScheduler(nil)

		g := NewDAG(scheduler)

		first := NewNode("computing_first_sum", func(ctx context.Context) error {
			firstComputation = 1 + 1
			return nil
		})
		g.AddNode(first)

		second := NewNode("computing_second_sum", func(ctx context.Context) error {
			secondComputation = firstComputation + 1
			return nil
		})
		g.AddNode(second)
		g.AddEdges(second, first)

		third := NewNode("computing_second_sum", func(ctx context.Context) error {
			thirdComputation = firstComputation + 2
			return nil
		})
		g.AddNode(third)
		g.AddEdges(third, first)

		final := NewNode("computing_final_sum", func(ctx context.Context) error {
			finalComputation = secondComputation + thirdComputation + 1
			return nil
		})
		g.AddNode(final)
		g.AddEdges(final, second, third)

		ctx := context.Background()

		err := g.Compute(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 8, finalComputation)
	})

	t.Run("error", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		g := NewDAG(scheduler)

		first := NewNode("computing_first_sum", func(ctx context.Context) error {
			return fmt.Errorf("This looks pretty bad!")
		})
		g.AddNode(first)

		second := NewNode("computing_second_sum", func(ctx context.Context) error {
			return nil
		})
		g.AddNode(second)
		g.AddEdges(second, first)

		ctx := context.Background()

		err := g.Compute(ctx)
		assert.Error(t, err)
		computationErr, ok := err.(*ComputationError)
		assert.True(t, ok)
		assert.Equal(t, "computing_first_sum", computationErr.NodeName)
	})
}
