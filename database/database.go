package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	dbInstance *sql.DB
	once       sync.Once
	initError  error
)

// GetDatabaseInstance シングルトンパターンでデータベース接続を提供
func GetDatabaseInstance(db Database) (*sql.DB, error) {
	once.Do(func() {
		dbInstance, initError = db.Connect()
	})
	return dbInstance, initError
}

type CockroachDB struct{}

func (db CockroachDB) Connect() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	databaseName := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	dbconf := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, databaseName, sslmode)
	dbConn, err := sql.Open("postgres", dbconf)
	if err != nil {
		return nil, err
	}

	dbConn.SetConnMaxLifetime(time.Minute * 3)
	dbConn.SetMaxOpenConns(10)
	dbConn.SetMaxIdleConns(10)

	if err := dbConn.Ping(); err != nil {
		return nil, err
	}
	return dbConn, nil
}
