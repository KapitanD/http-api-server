package model

import (
	"testing"
	"time"
)

// TestUser ...
func TestUser(t *testing.T) *User {
	return &User{
		Email:    "user@example.org",
		Password: "password",
	}
}

// TestNote ...
func TestNote(t *testing.T) *Note {
	return &Note{
		Header:    "header",
		Body:      "body",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
