package internaldto

import "github.com/stackql/stackql/internal/stackql/typing"

var (
	_ ViewDTO = &standardViewDTO{}
)

func NewViewDTO(viewName, rawViewQuery string) ViewDTO {
	return &standardViewDTO{
		viewName:     viewName,
		rawViewQuery: rawViewQuery,
	}
}

type ViewDTO interface {
	GetRawQuery() string
	GetName() string
	IsMaterialized() bool
	GetColumns() []typing.RelationalColumn
	SetColumns(columns []typing.RelationalColumn)
}

type standardViewDTO struct {
	rawViewQuery string
	viewName     string
	columns      []typing.RelationalColumn
}

func (v *standardViewDTO) GetRawQuery() string {
	return v.rawViewQuery
}

func (v *standardViewDTO) GetName() string {
	return v.viewName
}

func (v *standardViewDTO) IsMaterialized() bool {
	return false
}

func (v *standardViewDTO) GetColumns() []typing.RelationalColumn {
	return v.columns
}

func (v *standardViewDTO) SetColumns(columns []typing.RelationalColumn) {
	v.columns = columns
}
