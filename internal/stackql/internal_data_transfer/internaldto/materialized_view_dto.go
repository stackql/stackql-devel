package internaldto

import (
	"github.com/stackql/stackql/internal/stackql/typing"
)

var (
	_ ViewDTO = &standardMaterializedViewDTO{}
)

func NewMaterializedViewDTO(viewName, rawViewQuery string) ViewDTO {
	return &standardMaterializedViewDTO{
		viewName:     viewName,
		rawViewQuery: rawViewQuery,
	}
}

type standardMaterializedViewDTO struct {
	rawViewQuery string
	viewName     string
	columns      []typing.RelationalColumn
}

func (v *standardMaterializedViewDTO) GetRawQuery() string {
	return v.rawViewQuery
}

func (v *standardMaterializedViewDTO) GetName() string {
	return v.viewName
}

func (v *standardMaterializedViewDTO) IsMaterialized() bool {
	return true
}

func (v *standardMaterializedViewDTO) GetColumns() []typing.RelationalColumn {
	return v.columns
}

func (v *standardMaterializedViewDTO) SetColumns(columns []typing.RelationalColumn) {
	v.columns = columns
}
