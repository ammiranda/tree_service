package models

// Node represents a single node in the tree
type Node struct {
	ID       int64   `json:"id"`
	Label    string  `json:"label" validate:"required"`
	Children []*Node `json:"children"`
}

// NewNode creates a new node with the given label
func NewNode(label string) *Node {
	return &Node{
		Label:    label,
		Children: make([]*Node, 0),
	}
}

// AddChild adds a child node to the current node
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
}
