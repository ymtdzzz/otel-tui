package table

type TableID string

const (
	TABLE_ID_TRACE TableID = "table_trace"
)

type UpdateTableMsg struct {
	ID TableID
}
