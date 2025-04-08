package tree

type TreeID string

const (
	TREE_ID_TRACE TreeID = "tree_trace"
)

type UpdateTreeMsg struct {
	ID   TreeID
	Root *TreeNode
}
