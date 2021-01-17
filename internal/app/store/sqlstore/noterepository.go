package sqlstore

import (
	"database/sql"
	"time"

	"github.com/KapitanD/http-api-server/internal/app/model"
	"github.com/KapitanD/http-api-server/internal/app/store"
)

// NoteRepository ...
type NoteRepository struct {
	store *Store
}

// Create ...
func (r *NoteRepository) Create(n *model.Note, u *model.User) error {
	if err := n.Validate(); err != nil {
		return err
	}

	n.AuthorID = u.ID

	return r.store.db.QueryRow(
		"INSERT INTO notes (author_id, header, body, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id;",
		u.ID,
		n.Header,
		n.Body,
		n.CreatedAt,
		n.UpdatedAt,
	).Scan(&n.ID)
}

// Update ...
func (r *NoteRepository) Update(id int, un *model.Note) error {
	if err := un.ValidateUpdate(); err != nil {
		return err
	}

	n, err := r.FindByID(id)
	if err != nil {
		return err
	}
	if un.Body != "" {
		n.Body = un.Body
	}
	if un.Header != "" {
		n.Header = un.Header
	}
	n.UpdatedAt = time.Now()
	return r.store.db.QueryRow(
		"UPDATE notes SET header=$1, body=$2, updated_at=$3 WHERE id=$4 RETURNING id;",
		n.Header,
		n.Body,
		n.UpdatedAt,
		id,
	).Scan(&id)
}

// Delete ...
func (r *NoteRepository) Delete(id int) error {
	_, err := r.store.db.Exec(
		"DELETE FROM notes WHERE id = $1;",
		id,
	)
	return err
}

// FindByUser ...
func (r *NoteRepository) FindByUser(u *model.User) ([]*model.Note, error) {
	rows, err := r.store.db.Query(
		"SELECT id, author_id, header, body, created_at, updated_at FROM notes WHERE author_id=$1",
		u.ID,
	)
	if err != nil {
		return nil, err
	}
	result := []*model.Note{}
	for rows.Next() {
		n := &model.Note{}
		if err := rows.Scan(
			&n.ID,
			&n.AuthorID,
			&n.Header,
			&n.Body,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			if err == sql.ErrNoRows {
				return nil, store.ErrRecordNotFound
			}
			return nil, err
		}
		result = append(result, n)
	}
	return result, nil
}

// FindByID ...
func (r *NoteRepository) FindByID(id int) (*model.Note, error) {
	n := &model.Note{}
	if err := r.store.db.QueryRow(
		"SELECT id, author_id, header, body, created_at, updated_at FROM notes WHERE id = $1",
		id,
	).Scan(
		&n.ID,
		&n.AuthorID,
		&n.Header,
		&n.Body,
		&n.CreatedAt,
		&n.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	return n, nil
}
