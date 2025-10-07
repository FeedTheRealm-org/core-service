package repositories

import (
	"context"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/jackc/pgx/v5"
)

type exampleRepository struct {
	conf *config.Config
	conn *pgx.Conn
}

func NewExampleRepository(conf *config.Config) (ExampleRepository, error) {
	conn, err := conf.Dbc.GetConnectionToDatabase()
	if err != nil {
		return nil, err
	}

	return &exampleRepository{
		conf: conf,
		conn: conn,
	}, nil
}

func (er *exampleRepository) GetExampleRecord() string {
	return "IM AUTH"
}

func (er *exampleRepository) GetSumQuery() int {
	var result int
	row := er.conn.QueryRow(context.Background(), "SELECT 1 + 1 AS result;")
	if err := row.Scan(&result); err != nil {
		return 0
	}
	return result
}
