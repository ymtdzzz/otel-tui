package tuiv2

const (
	MODAL_TYPE_TRACE = iota
	// MODAL_TYPE_METRIC
	// MODAL_TYPE_LOG
)

type SetTextModalMsg struct {
	Type int
	Text string
}
