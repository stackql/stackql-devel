package internaldto

var (
	_ MaterializedViewDTO = &standardMaterializedViewDTO{}
)

func NewMaterializedViewDTO(viewName, rawViewQuery string) MaterializedViewDTO {
	return &standardMaterializedViewDTO{
		viewName:     viewName,
		rawViewQuery: rawViewQuery,
	}
}

type MaterializedViewDTO interface {
	GetRawQuery() string
	GetName() string
}

type standardMaterializedViewDTO struct {
	rawViewQuery string
	viewName     string
}

func (v *standardMaterializedViewDTO) GetRawQuery() string {
	return v.rawViewQuery
}

func (v *standardMaterializedViewDTO) GetName() string {
	return v.viewName
}
