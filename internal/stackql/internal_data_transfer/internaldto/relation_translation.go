package internaldto

var (
	_ RelationDMLTranslation = &standardDMLTranslation{}
)

type RelationDMLTranslation interface {
	GetRawQuery() string
	// Analyze() error
	GetTranlatedDDL() string
	GetLoadDML() (string, bool)
}

func NewRelationDMLTranslation(
	rawQuery string,
	translatedDDL string,
	loadDML string,
) RelationDMLTranslation {
	return &standardDMLTranslation{
		rawQuery:      rawQuery,
		translatedDDL: translatedDDL,
		loadDML:       loadDML,
	}
}

type standardDMLTranslation struct {
	rawQuery      string
	translatedDDL string
	loadDML       string
}

func (s *standardDMLTranslation) GetRawQuery() string {
	return s.rawQuery
}

func (s *standardDMLTranslation) GetTranlatedDDL() string {
	return s.translatedDDL
}

func (s *standardDMLTranslation) GetLoadDML() (string, bool) {
	return s.loadDML, s.loadDML != ""
}

func (s *standardDMLTranslation) Analyze() error {
	return nil
}
