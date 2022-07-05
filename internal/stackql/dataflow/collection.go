package dataflow

type DataFlowCollection interface {
	AddEdge(e DataFlowEdge) error
}

func NewStandardDataFlowCollection() DataFlowCollection {
	return &StandardDataFlowCollection{}
}

type StandardDataFlowCollection struct {
	vertices []DataFlowVertex
	edges    []DataFlowEdge
}

func (dc *StandardDataFlowCollection) AddEdge(e DataFlowEdge) error {
	dc.vertices = append(dc.vertices, nil)
	dc.edges = append(dc.edges, e)
	return nil
}
