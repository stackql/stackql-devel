package primitivegraph

import (
	"fmt"
	"infraql/internal/iql/drm"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/primitive"

	"gonum.org/v1/gonum/graph"

	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type PrimitiveGraph struct {
	g      *simple.WeightedDirectedGraph
	sorted []graph.Node
}

func (pg *PrimitiveGraph) Execute(ctx primitive.IPrimitiveCtx) dto.ExecutorOutput {
	var output dto.ExecutorOutput = dto.NewExecutorOutput(nil, nil, nil, nil, fmt.Errorf("empty execution graph"))
	for _, node := range pg.sorted {
		switch node := node.(type) {
		case PrimitiveNode:
			output = node.Primitive.Execute(ctx)
			destinationNodes := pg.g.From(node.ID())
			for {
				if !destinationNodes.Next() {
					break
				}
				fromNode := destinationNodes.Node()
				switch fromNode := fromNode.(type) {
				case PrimitiveNode:
					fromNode.Primitive.IncidentData(node.ID(), output)
				}
			}
		default:
			dto.NewExecutorOutput(nil, nil, nil, nil, fmt.Errorf("unknown execution primitive type: '%T'", node))
		}
	}
	return output
}

func (pg *PrimitiveGraph) GetPreparedStatementContext() *drm.PreparedStatementCtx {
	return nil
}

func (pg *PrimitiveGraph) SetTxnId(id int) {
	nodes := pg.g.Nodes()
	for {
		if !nodes.Next() {
			return
		}
		node := nodes.Node()
		switch node := node.(type) {
		case PrimitiveNode:
			node.Primitive.SetTxnId(id)
		}
	}
}

func (pg *PrimitiveGraph) Optimise() error {
	var err error
	pg.sorted, err = topo.Sort(pg.g)
	return err
}

func (pg *PrimitiveGraph) IncidentData(fromId int64, input dto.ExecutorOutput) error {
	return nil
}

func (pr *PrimitiveGraph) SetInputAlias(alias string, id int64) error {
	return nil
}

func SortPlan(g *PrimitiveGraph) (sorted []graph.Node, err error) {
	return topo.Sort(g.g)
}

type PrimitiveNode struct {
	Primitive primitive.IPrimitive
	id        int64
}

func (pg *PrimitiveGraph) CreatePrimitiveNode(pr primitive.IPrimitive) PrimitiveNode {
	nn := pg.g.NewNode()
	node := PrimitiveNode{
		Primitive: pr,
		id:        nn.ID(),
	}
	pg.g.AddNode(node)
	return node
}

func (pn PrimitiveNode) ID() int64 {
	return pn.id
}

func (pn PrimitiveNode) SetInputAlias(alias string, id int64) error {
	return pn.Primitive.SetInputAlias(alias, id)
}

func NewPrimitiveGraph() *PrimitiveGraph {
	return &PrimitiveGraph{g: simple.NewWeightedDirectedGraph(0.0, 0.0)}
}

func (pg *PrimitiveGraph) NewDependency(from PrimitiveNode, to PrimitiveNode, weight float64) {
	e := pg.g.NewWeightedEdge(from, to, weight)
	pg.g.SetWeightedEdge(e)
}

func SingletonPrimitiveGraph(pr primitive.IPrimitive) *PrimitiveGraph {
	gr := NewPrimitiveGraph()
	gr.CreatePrimitiveNode(pr)
	return gr
}
