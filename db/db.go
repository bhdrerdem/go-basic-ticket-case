package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type DatabaseInterface interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	BeginTransaction() (*sql.Tx, error)
	Ping() error
	Close() error
	IsHealthy() bool
}

var DB Database

type Database struct {
	client    *sql.DB
	isHealthy bool
}

func InitDB() {
	err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	go startDBHealthCheck()
}

func connectDB() error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	var err error
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	DB.client, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %v", err)
	}

	log.Println("Connected to database!!")
	return nil
}

func startDBHealthCheck() {
	DB.isHealthy = true
	ticker := time.NewTicker(10000 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if err := DB.Ping(); err != nil {
			log.Printf("Database ping failed: %v", err)
			DB.isHealthy = false

			if err := connectDB(); err != nil {
				log.Printf("Failed to reconnect to DB: %v", err)
			} else {
				DB.isHealthy = true
				log.Println("Database reconnected")
			}
		}
	}
}

func (d *Database) IsHealthy() bool {
	return d.isHealthy
}

func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.client.QueryRow(query, args...)
}

func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.client.Exec(query, args...)
}

func (d *Database) BeginTransaction() (*sql.Tx, error) {
	return d.client.Begin()
}

func (d *Database) Ping() error {
	return d.client.Ping()
}

func (d *Database) Close() error {
	return d.client.Close()
}
