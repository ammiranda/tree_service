package models

import "github.com/go-playground/validator/v10"

// Node represents a single node in the tree
type Node struct {
	ID       int64   `json:"id"`
	Label    string  `json:"label" validate:"required"`
	Children []*Node `json:"children"`
}

// CreateNodeRequest represents the request body for creating a new node
type CreateNodeRequest struct {
	Label    string `json:"label" validate:"required,min=1,max=100"`
	ParentID int64  `json:"parentId" validate:"omitempty,gt=0"`
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

// Validate validates the CreateNodeRequest
func (r *CreateNodeRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
