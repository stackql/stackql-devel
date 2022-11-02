package relationaldto

var (
	_ RelationalTable = &standardRelationalTable{}
)

type RelationalTable interface {
	GetColumns() []RelationalColumn
	GetName() string
	PushBackColumn(RelationalColumn)
}

func NewRelationalTable(name string) RelationalTable {
	return &standardRelationalTable{
		name: name,
	}
}

type standardRelationalTable struct {
	name    string
	columns []RelationalColumn
}

func (rt *standardRelationalTable) GetName() string {
	return rt.name
}

func (rt *standardRelationalTable) GetColumns() []RelationalColumn {
	return rt.columns
}

func (rt *standardRelationalTable) PushBackColumn(col RelationalColumn) {
	rt.columns = append(rt.columns, col)
}
