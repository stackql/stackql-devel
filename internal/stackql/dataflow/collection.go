package dataflow

type DataFlowUnit interface {
	iDataFlowUnit()
}

type DataFlowCollection interface {
	AddEdge(e DataFlowEdge) error
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

func (dc *StandardDataFlowCollection) GetAllUnits() []DataFlowUnit {
	return nil
}
