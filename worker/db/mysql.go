// Package db provides functions for interacting with MySQL
package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	log "worker-service/utils"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLClient struct {
	DB *sql.DB
}

func (db *MySQLClient) Connect() error {
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOST")
	dbname := os.Getenv("MYSQL_DBNAME")
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	log.FailOnError(err, "Failed to convert MYSQL_PORT to int")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, dbname)
	fmt.Printf("Connecting to MySQL: %s\n", dsn)

	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("Error connecting to MySQL: %s", err)
	}

	database.SetConnMaxLifetime(time.Minute * 3)
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(10)

	if err := database.Ping(); err != nil {
		log.FailOnError(err, "Error ping DB")
	}

	db.DB = database
	fmt.Println("✅ Connected to MySQL")

	return nil
}

func (db *MySQLClient) CloseConnection() {
	db.DB.Close()
	fmt.Println("✅ Closed connection to MySQL")
}
