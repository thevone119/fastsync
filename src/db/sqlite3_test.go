package db

import (
	"database/sql"
	"testing"
)


func TestSqllite3(t *testing.T) {

	db, err := sql.Open("sqlite3", "./foo.db")


}
