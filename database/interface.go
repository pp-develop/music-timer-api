package database

import (
	"database/sql"
)

type Database interface {
	Connect() (*sql.DB, error)
}
