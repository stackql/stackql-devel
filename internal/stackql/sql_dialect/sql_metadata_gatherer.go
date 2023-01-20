package sql_dialect

type SQLMetadataGatherer interface {
	// query string
}

type standardSQLMetadataGatherer struct {
	query string
}

func (mg *standardSQLMetadataGatherer) Get() {}
