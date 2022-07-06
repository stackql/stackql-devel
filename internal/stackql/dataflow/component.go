package dataflow

import (
	"fmt"

	"gonum.org/v1/gonum/graph"
)

type DataFlowWeaklyConnectedComponent interface {
	DataFlowUnit
	AddEdge(DataFlowEdge)
	Analyze() error
	PushBack(DataFlowVertex)
}

type StandardDataFlowWeaklyConnectedComponent struct {
	idsVisited   map[int64]struct{}
	collection   *StandardDataFlowCollection
	root         graph.Node
	orderedNodes []graph.Node
	edges        []graph.Edge
}

func NewStandardDataFlowWeaklyConnectedComponent(
	collection *StandardDataFlowCollection,
	root graph.Node,
) DataFlowWeaklyConnectedComponent {
	return &StandardDataFlowWeaklyConnectedComponent{
		collection: collection,
		root:       root,
		idsVisited: map[int64]struct{}{
			root.ID(): {},
		},
	}
}

func (wc *StandardDataFlowWeaklyConnectedComponent) Analyze() error {
	for _, node := range wc.collection.sorted {
		incidentNodes := wc.collection.g.From(node.ID())
		for {
			itemPresent := incidentNodes.Next()
			if !itemPresent {
				break
			}
			fromNode := incidentNodes.Node()
			_, ok := wc.idsVisited[fromNode.ID()]
			if ok {
				wc.orderedNodes = append(wc.orderedNodes, node)
				wc.idsVisited[node.ID()] = struct{}{}
				incidentEdge := wc.collection.g.WeightedEdge(fromNode.ID(), node.ID())
				if incidentEdge == nil {
					return fmt.Errorf("found nil edge in data flow graph")
				}
				wc.edges = append(wc.edges, incidentEdge)
			}
		}
	}
	return nil
}

func (wc *StandardDataFlowWeaklyConnectedComponent) AddEdge(e DataFlowEdge) {
	wc.edges = append(wc.edges, e)
}

func (wc *StandardDataFlowWeaklyConnectedComponent) PushBack(v DataFlowVertex) {
	wc.orderedNodes = append(wc.orderedNodes, v)
}

func (wc *StandardDataFlowWeaklyConnectedComponent) iDataFlowUnit() {}
