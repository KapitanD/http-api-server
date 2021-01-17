package store

import "github.com/KapitanD/http-api-server/internal/app/model"

// UserRepository ...
type UserRepository interface {
	Create(*model.User) error
	FindByEmail(string) (*model.User, error)
	Find(int) (*model.User, error)
}

// NoteRepository ...
type NoteRepository interface {
	Create(*model.Note, *model.User) error
	Update(int, *model.Note) error
	Delete(int) error
	FindByUser(*model.User) ([]*model.Note, error)
	FindByID(int) (*model.Note, error)
}
