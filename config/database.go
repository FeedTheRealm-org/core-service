package config

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DatabaseConfig struct {
	username string
	password string
	host     string
	port     int
	database string
}

func NewDatabaseConfig(username string, password string, host string, port int, database string) *DatabaseConfig {
	return &DatabaseConfig{
		username: username,
		password: password,
		host:     host,
		port:     port,
		database: database,
	}
}

func (dbc *DatabaseConfig) GetConnectionToDatabase() (*pgx.Conn, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", dbc.username, dbc.password, dbc.host, dbc.port, dbc.database)
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
