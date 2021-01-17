package model_test

import (
	"testing"

	"github.com/KapitanD/http-api-server/internal/app/model"
	"github.com/stretchr/testify/assert"
)

func TestNote_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		n       func() *model.Note
		isValid bool
	}{
		{
			name: "valid note",
			n: func() *model.Note {
				return model.TestNote(t)
			},
			isValid: true,
		},
		{
			name: "empty header",
			n: func() *model.Note {
				n := model.TestNote(t)
				n.Header = ""
				return n
			},
			isValid: false,
		},
		{
			name: "empty body",
			n: func() *model.Note {
				n := model.TestNote(t)
				n.Body = ""
				return n
			},
			isValid: false,
		},
		{
			name: "large header",
			n: func() *model.Note {
				n := model.TestNote(t)
				n.Header = `
				headerheaderheaderheaderheaderheaderheaderheaderheaderheaderheader
				headerheaderheaderheaderheaderheaderheaderheaderheaderheaderheader
				headerheaderheaderheaderheaderheaderheaderheaderheaderheaderheader
				headerheaderheaderheaderheaderheaderheaderheaderheaderheaderheader
				`
				return n
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		if tc.isValid {
			assert.NoError(t, tc.n().Validate())
		} else {
			assert.Error(t, tc.n().Validate())
		}
	}
}
