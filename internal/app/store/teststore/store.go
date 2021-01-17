package teststore

import (
	"github.com/KapitanD/http-api-server/internal/app/model"
	"github.com/KapitanD/http-api-server/internal/app/store"
)

// Store ...
type Store struct {
	userRepository *UserRepository
	noteRepository *NoteRepository
}

// New ...
func New() *Store {
	return &Store{}
}

// User ...
func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
		users: make(map[int]*model.User),
	}

	return s.userRepository
}

// Notes ...
func (s *Store) Notes() store.NoteRepository {
	if s.noteRepository != nil {
		return s.noteRepository
	}

	s.noteRepository = &NoteRepository{
		store: s,
		notes: make(map[int]*model.Note),
	}

	return s.noteRepository
}
