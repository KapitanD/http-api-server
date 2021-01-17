package sqlstore

import (
	"database/sql"

	"github.com/KapitanD/http-api-server/internal/app/store"
	_ "github.com/lib/pq" // use pg driver
)

// Store ...
type Store struct {
	db             *sql.DB
	userRepository *UserRepository
	noteRepository *NoteRepository
}

// New ...
func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// User ...
func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
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
	}

	return s.noteRepository
}
