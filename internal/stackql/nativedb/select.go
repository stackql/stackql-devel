package nativedb

type Select interface {
	GetColumns() []Column
}

func NewSelect(columns []Column) Select {
	return &StandardSelect{
		columns: columns,
	}
}

type StandardSelect struct {
	columns []Column
}

func (sc *StandardSelect) GetColumns() []Column {
	return sc.columns
}
