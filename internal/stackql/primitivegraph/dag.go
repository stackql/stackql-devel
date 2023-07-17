package primitivegraph

import (
	"context"
	"fmt"

	"github.com/stackql/stackql/internal/stackql/acid/binlog"
	"github.com/stackql/stackql/internal/stackql/acid/operation"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/primitive"

	"gonum.org/v1/gonum/graph"

	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"

	"golang.org/x/sync/errgroup"
)

var (
	_ PrimitiveGraph       = (*standardPrimitiveGraph)(nil)
	_ primitive.IPrimitive = (*standardPrimitiveGraph)(nil)
)

type PrimitiveGraph interface {
	primitive.IPrimitive
	AddTxnControlCounters(t internaldto.TxnControlCounters)
	ContainsIndirect() bool
	CreatePrimitiveNode(pr primitive.IPrimitive) PrimitiveNode
	Execute(ctx primitive.IPrimitiveCtx) internaldto.ExecutorOutput
	GetInputFromAlias(string) (internaldto.ExecutorOutput, bool)
	IncidentData(fromID int64, input internaldto.ExecutorOutput) error
	GetTxnControlCounterSlice() []internaldto.TxnControlCounters
	NewDependency(from PrimitiveNode, to PrimitiveNode, weight float64)
	Optimise() error
	SetContainsIndirect(containsView bool)
	SetExecutor(func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput) error
	SetInputAlias(alias string, id int64) error
	SetTxnID(id int)
	Sort() (sorted []graph.Node, err error)
	Size() int
}

type standardPrimitiveGraph struct {
	g                      *simple.WeightedDirectedGraph
	sorted                 []graph.Node
	txnControlCounterSlice []internaldto.TxnControlCounters
	errGroup               *errgroup.Group
	errGroupCtx            context.Context
	containsView           bool
}

func (pg *standardPrimitiveGraph) Size() int {
	return pg.g.Nodes().Len()
}

func (pg *standardPrimitiveGraph) IsReadOnly() bool {
	nodes := pg.g.Nodes()
	for nodes.Next() {
		node := nodes.Node()
		primNode, isPrimNode := node.(PrimitiveNode)
		if !isPrimNode {
			continue
		}
		if !primNode.GetOperation().IsReadOnly() {
			return false
		}
	}
	return true
}

func (pg *standardPrimitiveGraph) SetRedoLog(binlog.LogEntry) {
}

func (pg *standardPrimitiveGraph) SetUndoLog(binlog.LogEntry) {
}

func (pg *standardPrimitiveGraph) GetRedoLog() (binlog.LogEntry, bool) {
	return nil, false
}

func (pg *standardPrimitiveGraph) GetUndoLog() (binlog.LogEntry, bool) {
	rv := binlog.NewSimpleLogEntry(nil, nil)
	for _, node := range pg.sorted {
		primNode, isPrimNode := node.(PrimitiveNode)
		if !isPrimNode {
			continue
		}
		op := primNode.GetOperation()
		undoLog, undoLogExists := op.GetUndoLog()
		if undoLogExists && undoLog != nil {
			rv.AppendRaw(undoLog.GetRaw())
			for _, h := range undoLog.GetHumanReadable() {
				rv.AppendHumanReadable(h)
			}
		}
	}
	return nil, false
}

func (pg *standardPrimitiveGraph) AddTxnControlCounters(t internaldto.TxnControlCounters) {
	pg.txnControlCounterSlice = append(pg.txnControlCounterSlice, t)
}

func (pg *standardPrimitiveGraph) GetTxnControlCounterSlice() []internaldto.TxnControlCounters {
	return pg.txnControlCounterSlice
}

func (pg *standardPrimitiveGraph) SetExecutor(func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput) error {
	return fmt.Errorf("pass through primitive does not support SetExecutor()")
}

func (pg *standardPrimitiveGraph) ContainsIndirect() bool {
	return pg.containsView
}

func (pg *standardPrimitiveGraph) SetContainsIndirect(containsView bool) {
	pg.containsView = containsView
}

func (pg *standardPrimitiveGraph) GetInputFromAlias(string) (internaldto.ExecutorOutput, bool) {
	var rv internaldto.ExecutorOutput
	return rv, false
}

// After each query execution, the graph needs to be reset.
// This is so that cached queries can be re-executed.
func (pg *standardPrimitiveGraph) reset() {
	for _, node := range pg.sorted {
		switch node := node.(type) { //nolint:gocritic // acceptable
		case PrimitiveNode:
			select {
			case <-node.IsDone():
			default:
			}
		}
	}
}

// Execute() is the entry point for the execution of the graph.
// It is responsible for executing the graph in a topological order.
// This particular implementation:
//   - Uses the errgroup package to execute the graph in parallel.
//   - Blocks on any node that has a dependency that has not been executed.
//
//nolint:gocognit // inherent complexity
func (pg *standardPrimitiveGraph) Execute(ctx primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
	// Reset the graph.
	// Absolutely necessary for re-execution
	defer pg.reset()
	//nolint:stylecheck // prefer declarative
	var output internaldto.ExecutorOutput = internaldto.NewExecutorOutput(
		nil, nil, nil, nil, fmt.Errorf("empty execution graph"))
	for _, node := range pg.sorted {
		outChan := make(chan internaldto.ExecutorOutput, 1)
		switch node := node.(type) {
		case PrimitiveNode:
			incidentNodes := pg.g.To(node.ID())
			for {
				hasNext := incidentNodes.Next()
				if !hasNext {
					break
				}
				incidentNode := incidentNodes.Node()
				switch incidentNode := incidentNode.(type) {
				case PrimitiveNode:
					// await completion of the incident node
					// and replenish the IsDone() channel
					incidentNode.SetIsDone(<-incidentNode.IsDone())
				default:
					return internaldto.NewExecutorOutput(
						nil, nil, nil, nil,
						fmt.Errorf("unknown execution primitive type: '%T'", incidentNode))
				}
			}
			pg.errGroup.Go(
				func() error {
					output := node.GetOperation().Execute(ctx) //nolint:govet // intentional
					outChan <- output
					close(outChan)
					return output.GetError()
				},
			)
			destinationNodes := pg.g.From(node.ID())
			output = <-outChan
			for {
				if !destinationNodes.Next() {
					break
				}
				fromNode := destinationNodes.Node()
				switch fromNode := fromNode.(type) { //nolint:gocritic // acceptable
				case PrimitiveNode:
					op := fromNode.GetOperation()
					op.IncidentData(node.ID(), output) //nolint:errcheck // TODO: consider design options
				}
			}
			node.SetIsDone(true)
		default:
			return internaldto.NewExecutorOutput(nil, nil, nil, nil, fmt.Errorf("unknown execution primitive type: '%T'", node))
		}
	}
	if err := pg.errGroup.Wait(); err != nil {
		undoLog, _ := output.GetUndoLog()
		return internaldto.NewExecutorOutput(nil, nil, nil, nil, err).WithUndoLog(undoLog)
	}
	return output
}

func (pg *standardPrimitiveGraph) SetTxnID(id int) {
	nodes := pg.g.Nodes()
	for {
		if !nodes.Next() {
			return
		}
		node := nodes.Node()
		switch node := node.(type) { //nolint:gocritic // acceptable
		case PrimitiveNode:
			node.GetOperation().SetTxnID(id)
		}
	}
}

func (pg *standardPrimitiveGraph) Optimise() error {
	var err error
	pg.sorted, err = topo.Sort(pg.g)
	return err
}

//nolint:revive // future proofing
func (pg *standardPrimitiveGraph) IncidentData(fromID int64, input internaldto.ExecutorOutput) error {
	return nil
}

//nolint:revive // future proofing
func (pg *standardPrimitiveGraph) SetInputAlias(alias string, id int64) error {
	return nil
}

func (pg *standardPrimitiveGraph) Sort() ([]graph.Node, error) {
	return topo.Sort(pg.g)
}

func SortPlan(pg PrimitiveGraph) ([]graph.Node, error) {
	return pg.Sort()
}

type PrimitiveNode interface {
	GetOperation() operation.Operation
	ID() int64
	IsDone() chan (bool)
	GetError() (error, bool)
	SetError(error)
	SetInputAlias(alias string, id int64) error
	SetIsDone(bool)
}

type standardPrimitiveNode struct {
	op     operation.Operation
	id     int64
	isDone chan bool
	err    error
}

func (pg *standardPrimitiveGraph) CreatePrimitiveNode(pr primitive.IPrimitive) PrimitiveNode {
	nn := pg.g.NewNode()
	node := &standardPrimitiveNode{
		op:     operation.NewReversibleOperation(pr),
		id:     nn.ID(),
		isDone: make(chan bool, 1),
	}
	pg.g.AddNode(node)
	return node
}

func (pn *standardPrimitiveNode) ID() int64 {
	return pn.id
}

//nolint:revive // TODO: consider API change
func (pn *standardPrimitiveNode) GetError() (error, bool) {
	return pn.err, pn.err != nil
}

func (pn *standardPrimitiveNode) GetOperation() operation.Operation {
	return pn.op
}

func (pn *standardPrimitiveNode) IsDone() chan bool {
	return pn.isDone
}

func (pn *standardPrimitiveNode) SetInputAlias(alias string, id int64) error {
	op := pn.GetOperation()
	return op.SetInputAlias(alias, id)
}

func (pn *standardPrimitiveNode) SetIsDone(isDone bool) {
	pn.isDone <- isDone
}

func (pn *standardPrimitiveNode) SetError(err error) {
	pn.err = err
}

func newPrimitiveGraph(concurrencyLimit int) PrimitiveGraph {
	eg, egCtx := errgroup.WithContext(context.Background())
	eg.SetLimit(concurrencyLimit)
	return &standardPrimitiveGraph{
		g:           simple.NewWeightedDirectedGraph(0.0, 0.0),
		errGroup:    eg,
		errGroupCtx: egCtx,
	}
}

func (pg *standardPrimitiveGraph) NewDependency(from PrimitiveNode, to PrimitiveNode, weight float64) {
	e := pg.g.NewWeightedEdge(from, to, weight)
	pg.g.SetWeightedEdge(e)
}

var (
	_ PrimitiveGraphHolder = (*standardPrimitiveGraphHolder)(nil)
)

//nolint:revive // acceptable nomenclature
type PrimitiveGraphHolder interface {
	AddInverseTxnControlCounters(t internaldto.TxnControlCounters)
	AddTxnControlCounters(t internaldto.TxnControlCounters)
	ContainsIndirect() bool
	CreateInversePrimitiveNode(pr primitive.IPrimitive) PrimitiveNode
	CreatePrimitiveNode(pr primitive.IPrimitive) PrimitiveNode
	GetInversePrimitiveGraph() PrimitiveGraph
	GetInverseTxnControlCounterSlice() []internaldto.TxnControlCounters
	GetPrimitiveGraph() PrimitiveGraph
	GetTxnControlCounterSlice() []internaldto.TxnControlCounters
	InverseContainsIndirect() bool
	NewInverseDependency(from PrimitiveNode, to PrimitiveNode, weight float64)
	NewDependency(from PrimitiveNode, to PrimitiveNode, weight float64)
	SetContainsIndirect(bool)
	SetInverseContainsIndirect(bool)
	SetTxnID(int)
	SetInverseTxnID(int)
}

type standardPrimitiveGraphHolder struct {
	pg  PrimitiveGraph
	ipg PrimitiveGraph
}

func (pgh *standardPrimitiveGraphHolder) GetPrimitiveGraph() PrimitiveGraph {
	return pgh.pg
}

func (pgh *standardPrimitiveGraphHolder) SetTxnID(txnID int) {
	pgh.pg.SetTxnID(txnID)
}

func (pgh *standardPrimitiveGraphHolder) SetInverseTxnID(txnID int) {
	pgh.ipg.SetTxnID(txnID)
}

func (pgh *standardPrimitiveGraphHolder) GetInversePrimitiveGraph() PrimitiveGraph {
	return pgh.ipg
}

func (pgh *standardPrimitiveGraphHolder) AddTxnControlCounters(t internaldto.TxnControlCounters) {
	pgh.pg.AddTxnControlCounters(t)
}

func (pgh *standardPrimitiveGraphHolder) AddInverseTxnControlCounters(t internaldto.TxnControlCounters) {
	pgh.ipg.AddTxnControlCounters(t)
}

func (pgh *standardPrimitiveGraphHolder) CreatePrimitiveNode(pr primitive.IPrimitive) PrimitiveNode {
	return pgh.pg.CreatePrimitiveNode(pr)
}

func (pgh *standardPrimitiveGraphHolder) CreateInversePrimitiveNode(pr primitive.IPrimitive) PrimitiveNode {
	return pgh.ipg.CreatePrimitiveNode(pr)
}

func (pgh *standardPrimitiveGraphHolder) NewDependency(from PrimitiveNode, to PrimitiveNode, weight float64) {
	pgh.pg.NewDependency(from, to, weight)
}

func (pgh *standardPrimitiveGraphHolder) NewInverseDependency(from PrimitiveNode, to PrimitiveNode, weight float64) {
	pgh.ipg.NewDependency(to, from, weight)
}

func (pgh *standardPrimitiveGraphHolder) SetContainsIndirect(containsView bool) {
	pgh.pg.SetContainsIndirect(containsView)
}

func (pgh *standardPrimitiveGraphHolder) ContainsIndirect() bool {
	return pgh.pg.ContainsIndirect()
}

func (pgh *standardPrimitiveGraphHolder) InverseContainsIndirect() bool {
	return pgh.ipg.ContainsIndirect()
}

func (pgh *standardPrimitiveGraphHolder) GetTxnControlCounterSlice() []internaldto.TxnControlCounters {
	return pgh.pg.GetTxnControlCounterSlice()
}

func (pgh *standardPrimitiveGraphHolder) GetInverseTxnControlCounterSlice() []internaldto.TxnControlCounters {
	return pgh.ipg.GetTxnControlCounterSlice()
}

func (pgh *standardPrimitiveGraphHolder) SetInverseContainsIndirect(containsView bool) {
	pgh.pg.SetContainsIndirect(containsView)
}

func NewPrimitiveGraphHolder(concurrencyLimit int) PrimitiveGraphHolder {
	pg := newPrimitiveGraph(concurrencyLimit)
	ipg := newPrimitiveGraph(concurrencyLimit)
	return &standardPrimitiveGraphHolder{
		pg:  pg,
		ipg: ipg,
	}
}
