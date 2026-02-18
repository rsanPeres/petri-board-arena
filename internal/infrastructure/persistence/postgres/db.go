package postgres

import (
	"database/sql"
)

type DB struct {
	SQL *sql.DB
}
