package graph

import (
	"context"
	"math/rand"
	"sync"
)

type NodeId int64

type Task func(ctx context.Context) error

type Node struct {
	name string
	id   NodeId
	task Task

	finished bool
}

func NewNode(name string, task Task) *Node {
	return &Node{
		name: name,
		id:   NodeId(rand.Int63()),
		task: task,
	}
}

type IGraph interface {
	AddNode(node *Node)
	AddEdges(node *Node, dependencies ...*Node)

	Compute(ctx context.Context) error
}

type nodeInformation struct {
	node *Node
	to   []NodeId // edges to other nodes
	from []NodeId // edges from other nodes
}

type DAG struct {
	scheduler IScheduler

	nodes map[NodeId]*nodeInformation
}

func NewDAG() *DAG {
	return &DAG{
		nodes: make(map[NodeId]*nodeInformation),
	}
}

func (g *DAG) AddNode(node *Node) {
	g.nodes[node.id] = &nodeInformation{
		node: node,
	}
}

func (g *DAG) AddEdges(to *Node, fromNodes ...*Node) {
	for _, from := range fromNodes {
		nodeInfo := g.nodes[to.id]
		nodeInfo.from = append(nodeInfo.from, from.id)

		nodeInfo = g.nodes[from.id]
		nodeInfo.to = append(nodeInfo.to, to.id)
	}
}

func (g *DAG) Compute(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rootIds := g.findRootNodes()
	for _, rootId := range rootIds {
		g.scheduleNode(ctx, rootId)
	}

	workResultCh := g.scheduler.Start()

	var computeError error

	shutdown := sync.Once{}
	for finishedWork := range workResultCh {
		if err := finishedWork.err; err != nil {
			// Call this only once as we don't want to resetting the error and
			// shutthing down the scheduler multiple times.
			shutdown.Do(func() {
				computeError = err
				cancel()

				// Shutdown the scheduler in a goroutine as it's a blocking call and we don't want to block
				// reading from the channel.
				go g.scheduler.Shutdown()
			})
		}

		// Don't schedule new tasks if we have failed.
		if computeError != nil {
			continue
		}

		for _, nodeId := range g.nodes[NodeId(finishedWork.Id)].to {
			if g.isNodeSchedulable(nodeId) {
				g.scheduleNode(ctx, nodeId)
			}
		}
	}

	return computeError
}

// Root nodes will have zero `to` dependencies. I.e. there is no edge directed at them.
func (g *DAG) findRootNodes() []NodeId {
	roots := make([]NodeId, 0)
	for _, n := range g.nodes {
		if len(n.from) == 0 {
			roots = append(roots, n.node.id)
		}
	}

	return roots
}

func (g *DAG) isNodeSchedulable(nodeId NodeId) bool {
	for _, toNodeId := range g.nodes[nodeId].from {
		if !g.nodes[toNodeId].node.finished {
			return false
		}
	}

	return true
}

func (g *DAG) scheduleNode(ctx context.Context, nodeId NodeId) {
	node := g.nodes[nodeId].node
	g.scheduler.Schedule(int64(nodeId), func() error {
		return node.task(ctx)
	})
}
