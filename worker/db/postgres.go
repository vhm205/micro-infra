// Package db provides functions for interacting with MySQL
package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type PostgresClient struct {
	URL string
	DB  *sql.DB
}

func (db *PostgresClient) Connect() error {
	database, err := sql.Open("postgres", db.URL)
	if err != nil {
		return fmt.Errorf("Error connecting to Postgres: %s", err)
	}

	database.SetConnMaxLifetime(time.Minute * 3)
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(10)

	if err := database.Ping(); err != nil {
		log.Fatal("Error ping DB:", err)
	}

	db.DB = database
	fmt.Println("✅ Connected to Postgres")

	return nil
}

func (db *PostgresClient) CloseConnection() {
	db.DB.Close()
	fmt.Println("✅ Closed connection to Postgres")
}
