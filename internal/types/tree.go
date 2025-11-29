package types

// TreeNode represents a node in the key tree view
type TreeNode struct {
	Name         string
	FullPath     string
	Path         string // Alias for FullPath
	IsKey        bool
	IsExpandable bool
	Type         KeyType
	Children     []*TreeNode
	Expanded     bool
	ChildCount   int
	Count        int // Number of keys under this prefix
	Depth        int
}

// NewTreeNode creates a new tree node
func NewTreeNode(name, fullPath string, isKey bool) *TreeNode {
	return &TreeNode{
		Name:     name,
		FullPath: fullPath,
		IsKey:    isKey,
		Children: []*TreeNode{},
	}
}

// AddChild adds a child node
func (n *TreeNode) AddChild(child *TreeNode) {
	n.Children = append(n.Children, child)
	n.ChildCount++
}

// FindChild finds a child by name
func (n *TreeNode) FindChild(name string) *TreeNode {
	for _, child := range n.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// Toggle expands or collapses the node
func (n *TreeNode) Toggle() {
	n.Expanded = !n.Expanded
}

// GetDepth returns the depth of the node based on colons in path
func (n *TreeNode) GetDepth() int {
	depth := 0
	for _, c := range n.FullPath {
		if c == ':' {
			depth++
		}
	}
	return depth
}
