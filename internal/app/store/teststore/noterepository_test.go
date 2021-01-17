package teststore_test

import (
	"testing"

	"github.com/KapitanD/http-api-server/internal/app/model"
	"github.com/KapitanD/http-api-server/internal/app/store"
	"github.com/KapitanD/http-api-server/internal/app/store/teststore"
	"github.com/stretchr/testify/assert"
)

func TestNoteRepository_Create(t *testing.T) {

	s := teststore.New()
	u := model.TestUser(t)
	n := model.TestNote(t)
	s.User().Create(u)
	assert.NoError(t, s.Notes().Create(n, u))
	assert.NotNil(t, n)
}

func TestNoteRepository_Update(t *testing.T) {
	s := teststore.New()
	u := model.TestUser(t)
	n := model.TestNote(t)
	s.User().Create(u)
	s.Notes().Create(n, u)
	un := model.TestNote(t)
	un.Header = "some"
	un.Body = "changes"

	assert.NoError(t, s.Notes().Update(n.ID, un))
	assert.NotNil(t, n)
}

func TestNoteRepository_Delete(t *testing.T) {
	s := teststore.New()
	u := model.TestUser(t)
	n := model.TestNote(t)
	s.User().Create(u)
	s.Notes().Create(n, u)

	assert.NoError(t, s.Notes().Delete(n.ID))

	_, err := s.Notes().FindByID(n.ID)
	assert.EqualError(t, store.ErrRecordNotFound, err.Error())
}

func TestNoteRepository_FindByUser(t *testing.T) {
	s := teststore.New()
	u := model.TestUser(t)
	n := model.TestNote(t)

	s.User().Create(u)
	rn, err := s.Notes().FindByUser(u)
	assert.NoError(t, err)
	assert.Equal(t, []*model.Note{}, rn)

	s.Notes().Create(n, u)
	rn, err = s.Notes().FindByUser(u)
	assert.NoError(t, err)
	assert.Equal(t, len(rn), 1)
	// dont need to compare timestamp, other fields are content uniqness
	n.CreatedAt = rn[0].CreatedAt
	n.UpdatedAt = rn[0].UpdatedAt
	assert.Equal(t, []*model.Note{n}, rn)
}

func TestNoteRepository_FindByID(t *testing.T) {
	s := teststore.New()
	u := model.TestUser(t)
	n := model.TestNote(t)

	s.User().Create(u)
	_, err := s.Notes().FindByID(u.ID)
	assert.EqualError(t, store.ErrRecordNotFound, err.Error())

	s.Notes().Create(n, u)
	rn, err := s.Notes().FindByID(n.ID)
	assert.NoError(t, err)
	// dont need to compare timestamp, other fields are content uniqness
	n.CreatedAt = rn.CreatedAt
	n.UpdatedAt = rn.UpdatedAt
	assert.Equal(t, n, rn)
}
