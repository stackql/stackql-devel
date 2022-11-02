package relationaldto

var (
	_ RelationalColumn = &standardRelationalColumn{}
)

type RelationalColumn interface {
	GetName() string
	GetType() string
	GetWidth() int
}

func NewRelationalColumn(colName string, colType string, width int) RelationalColumn {
	return &standardRelationalColumn{
		colType: colType,
		colName: colName,
		width:   width,
	}
}

type standardRelationalColumn struct {
	colType string
	colName string
	width   int
}

func (rc *standardRelationalColumn) GetName() string {
	return rc.colName
}

func (rc *standardRelationalColumn) GetType() string {
	return rc.colType
}

func (rc *standardRelationalColumn) GetWidth() int {
	return rc.width
}
