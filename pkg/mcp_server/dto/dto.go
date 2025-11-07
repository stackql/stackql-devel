package dto

// Pure data transfer objects (no dependencies on server/backend code).

// GreetDTO carries a simple greeting payload.
type GreetDTO struct {
	Greeting string `json:"greeting"`
}

// ServerInfoDTO provides static server/environment information.
type ServerInfoDTO struct {
	Name       string `json:"name"`
	Info       string `json:"info"`
	IsReadOnly bool   `json:"is_read_only"`
}

// DBIdentityDTO represents the current database identity.
type DBIdentityDTO struct {
	Identity string `json:"identity"`
}

// QueryResultDTO represents a query response; Rows only populated for JSON format.
// Raw contains original textual result when not parsed; Warnings may include advisory messages
// (e.g. URL encode slashes in hierarchical resource keys).
type QueryResultDTO struct {
	Rows     []map[string]any `json:"rows,omitempty"`
	RowCount int              `json:"row_count"`
	Format   string           `json:"format"`
	Raw      string           `json:"raw,omitempty"`
	Warnings []string         `json:"warnings,omitempty"`
}

// SimpleRowsDTO wraps a plain rows array.
type SimpleRowsDTO struct {
	Rows []map[string]any `json:"rows"`
}

// SimpleTextDTO wraps a single text payload.
type SimpleTextDTO struct {
	Text string `json:"text"`
}
