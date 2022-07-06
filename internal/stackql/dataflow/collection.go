package dataflow

import (
	"fmt"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"vitess.io/vitess/go/vt/sqlparser"
)

type DataFlowUnit interface {
	iDataFlowUnit()
}

type DataFlowCollection interface {
	AddEdge(e DataFlowEdge) error
	AddVertex(v DataFlowVertex)
	GetAllUnits() ([]DataFlowUnit, error)
	GetNextID() int64
	InDegree(v DataFlowVertex) int
	OutDegree(v DataFlowVertex) int
	Sort() error
	Vertices() []DataFlowVertex
}

func NewStandardDataFlowCollection() DataFlowCollection {
	return &StandardDataFlowCollection{
		g:                     simple.NewWeightedDirectedGraph(0.0, 0.0),
		vertices:              make(map[DataFlowVertex]struct{}),
		verticesForTableExprs: make(map[sqlparser.TableExpr]struct{}),
	}
}

type StandardDataFlowCollection struct {
	maxId                 int64
	g                     *simple.WeightedDirectedGraph
	sorted                []graph.Node
	vertices              map[DataFlowVertex]struct{}
	verticesForTableExprs map[sqlparser.TableExpr]struct{}
	edges                 []DataFlowEdge
}

func (dc *StandardDataFlowCollection) GetNextID() int64 {
	dc.maxId++
	return dc.maxId
}

func (dc *StandardDataFlowCollection) AddEdge(e DataFlowEdge) error {
	dc.AddVertex(e.GetSource())
	dc.AddVertex(e.GetDest())
	dc.edges = append(dc.edges, e)
	dc.g.SetWeightedEdge(e)
	return nil
}

func (dc *StandardDataFlowCollection) AddVertex(v DataFlowVertex) {
	_, ok := dc.verticesForTableExprs[v.GetTableExpr()]
	if ok {
		return
	}
	dc.vertices[v] = struct{}{}
	dc.verticesForTableExprs[v.GetTableExpr()] = struct{}{}
	dc.g.AddNode(v)
}

func (dc *StandardDataFlowCollection) Sort() error {
	var err error
	dc.sorted, err = topo.Sort(dc.g)
	return err
}

func (dc *StandardDataFlowCollection) Vertices() []DataFlowVertex {
	var rv []DataFlowVertex
	for vert := range dc.vertices {
		rv = append(rv, vert)
	}
	return rv
}

func (dc *StandardDataFlowCollection) GetAllUnits() ([]DataFlowUnit, error) {
	var rv []DataFlowUnit
	for _, vert := range dc.sorted {
		switch vert := vert.(type) {
		case DataFlowUnit:
			rv = append(rv, vert)
		default:
			return nil, fmt.Errorf("cannot accomodate data flow unit of type = '%T'", vert)
		}
	}
	return rv, nil
}

func (dc *StandardDataFlowCollection) InDegree(v DataFlowVertex) int {
	inDegree := 0
	for _, e := range dc.edges {
		if e.GetDest() == v {
			inDegree++
		}
	}
	return inDegree
}

func (dc *StandardDataFlowCollection) OutDegree(v DataFlowVertex) int {
	outDegree := 0
	for _, e := range dc.edges {
		if e.GetSource() == v {
			outDegree++
		}
	}
	return outDegree
}
