package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)

func init() {
	var err error
	db, err = CockroachDBConnect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
}

func CockroachDBConnect() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	database_name := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	dbconf := "postgresql://" + user + ":" + password + "@" + host + ":" + port + "/" + database_name + "?sslmode=" + sslmode
	db, err := sql.Open("postgres", dbconf)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
