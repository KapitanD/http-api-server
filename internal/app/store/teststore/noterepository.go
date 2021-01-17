package teststore

import (
	"time"

	"github.com/KapitanD/http-api-server/internal/app/model"
	"github.com/KapitanD/http-api-server/internal/app/store"
)

// NoteRepository ...
type NoteRepository struct {
	store *Store
	notes map[int]*model.Note
}

// Create ...
func (r *NoteRepository) Create(n *model.Note, u *model.User) error {
	if err := n.Validate(); err != nil {
		return err
	}

	n.AuthorID = u.ID

	n.ID = len(r.notes) + 1
	r.notes[n.ID] = n

	return nil
}

// Update ...
func (r *NoteRepository) Update(id int, un *model.Note) error {
	if err := un.Validate(); err != nil {
		return err
	}
	n, ok := r.notes[id]
	if !ok {
		return store.ErrRecordNotFound
	}
	n.Body = un.Body
	n.Header = un.Header
	n.UpdatedAt = time.Now()
	return nil
}

// Delete ...
func (r *NoteRepository) Delete(id int) error {
	delete(r.notes, id)
	return nil
}

// FindByUser ...
func (r *NoteRepository) FindByUser(u *model.User) ([]*model.Note, error) {
	result := []*model.Note{}
	for _, n := range r.notes {
		if n.AuthorID == u.ID {
			result = append(result, n)
		}
	}
	return result, nil
}

// FindByID ...
func (r *NoteRepository) FindByID(id int) (*model.Note, error) {
	n, ok := r.notes[id]
	if !ok {
		return nil, store.ErrRecordNotFound
	}

	return n, nil
}
