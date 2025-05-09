package models

import (
	"github.com/go-playground/validator/v10"
)

// CreateNodeRequest represents the request body for creating a node
type CreateNodeRequest struct {
	Label    string `json:"label" validate:"required,min=1,max=100"`
	ParentID int64  `json:"parentId" validate:"omitempty,gt=0"`
}

// UpdateNodeRequest represents the request body for updating a node
type UpdateNodeRequest struct {
	Label    string `json:"label" validate:"required,min=1,max=100"`
	ParentID *int64 `json:"parentId,omitempty" validate:"omitempty,gt=0"`
}

// Validate validates the create node request
func (r *CreateNodeRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate validates the update node request
func (r *UpdateNodeRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
