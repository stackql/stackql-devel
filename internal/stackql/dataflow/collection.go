package dataflow

type DataFlowUnit interface {
	iDataFlowUnit()
}

type DataFlowCollection interface {
	AddEdge(e DataFlowEdge) error
	AddVertex(v DataFlowVertex)
	GetAllUnits() []DataFlowUnit
	InDegree(v DataFlowVertex) int
	OutDegree(v DataFlowVertex) int
}

func NewStandardDataFlowCollection() DataFlowCollection {
	return &StandardDataFlowCollection{
		vertices: make(map[DataFlowVertex]struct{}),
	}
}

type StandardDataFlowCollection struct {
	vertices map[DataFlowVertex]struct{}
	edges    []DataFlowEdge
}

func (dc *StandardDataFlowCollection) AddEdge(e DataFlowEdge) error {
	dc.vertices[e.GetSource()] = struct{}{}
	dc.vertices[e.GetDest()] = struct{}{}
	dc.edges = append(dc.edges, e)
	return nil
}

func (dc *StandardDataFlowCollection) AddVertex(v DataFlowVertex) {
	dc.vertices[v] = struct{}{}
}

func (dc *StandardDataFlowCollection) GetAllUnits() []DataFlowUnit {
	var rv []DataFlowUnit
	for vert := range dc.vertices {
		rv = append(rv, vert)
	}
	return rv
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
