package internaldto

import "github.com/stackql/stackql/internal/stackql/typing"

var (
	_ RelationDTO = &standardViewDTO{}
)

func NewViewDTO(viewName, rawViewQuery string) RelationDTO {
	return &standardViewDTO{
		viewName:     viewName,
		rawViewQuery: rawViewQuery,
	}
}

type RelationDTO interface {
	GetRawQuery() string
	GetName() string
	IsMaterialized() bool
	IsTable() bool
	GetNamespace() string
	GetColumns() []typing.RelationalColumn
	SetColumns(columns []typing.RelationalColumn)
	MatchOnParams(map[string]any) (RelationDTO, bool)
}

type standardViewDTO struct {
	rawViewQuery string
	viewName     string
	columns      []typing.RelationalColumn
}

func (v *standardViewDTO) GetRawQuery() string {
	return v.rawViewQuery
}

func (v *standardViewDTO) MatchOnParams(map[string]any) (RelationDTO, bool) {
	return v, true
}

func (v *standardViewDTO) GetName() string {
	return v.viewName
}

func (v *standardViewDTO) GetNamespace() string {
	return ""
}

func (v *standardViewDTO) IsMaterialized() bool {
	return false
}

func (v *standardViewDTO) IsTable() bool {
	return false
}

func (v *standardViewDTO) GetColumns() []typing.RelationalColumn {
	return v.columns
}

func (v *standardViewDTO) SetColumns(columns []typing.RelationalColumn) {
	v.columns = columns
}
