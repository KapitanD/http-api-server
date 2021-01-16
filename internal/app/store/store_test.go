package store_test

import (
	"os"
	"testing"
)

var (
	databaseURL string
)

func TestMain(m *testing.M) {
	databaseURL = os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=0.0.0.0 port=12345 dbname=restapi_test user=postgres password=example sslmode=disable"
	}

	os.Exit(m.Run())
}
