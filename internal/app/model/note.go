package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

// Note ...
type Note struct {
	ID        int       `json:"id"`
	AuthorID  int       `json:"author_id"`
	Header    string    `json:"header"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate ...
func (n *Note) Validate() error {
	return validation.ValidateStruct(
		n,
		validation.Field(&n.Header, validation.Required, validation.Length(1, 100)),
		validation.Field(&n.Body, validation.Required, validation.Length(1, 1000)),
	)
}

// ValidateUpdate ...
func (n *Note) ValidateUpdate() error {
	return validation.ValidateStruct(
		n,
		validation.Field(&n.Header, validation.Length(0, 100)),
		validation.Field(&n.Body, validation.Length(0, 1000)),
	)
}
