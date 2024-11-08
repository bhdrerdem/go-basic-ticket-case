package mocks

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
)

type MockDatabase struct {
	client *sql.DB
	mock   sqlmock.Sqlmock
}

func NewMockDatabase() (*MockDatabase, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	return &MockDatabase{client: db, mock: mock}, mock, nil
}

func (m *MockDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.client.QueryRow(query, args...)
}

func (m *MockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.client.Exec(query, args...)
}

func (m *MockDatabase) BeginTransaction() (*sql.Tx, error) {
	return m.client.Begin()
}

func (m *MockDatabase) Ping() error {
	return m.client.Ping()
}

func (m *MockDatabase) Close() error {
	return m.client.Close()
}

func (m *MockDatabase) IsHealthy() bool {
	return true
}
