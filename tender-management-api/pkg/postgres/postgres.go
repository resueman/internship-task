package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
)

type Postgres struct {
	Database   *sql.DB
	SqlBuilder squirrel.StatementBuilderType
}

func NewDB(url string) (*Postgres, error) {
	driver := "postgres"
	db, err := sql.Open(driver, url)
	if err != nil {
		return nil, fmt.Errorf("error while opening database with driver `%s` and url `%s`. %w", driver, url, err)
	}

	return &Postgres{
		Database:   db,
		SqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (p *Postgres) Close() error {
	if p.Database != nil {
		err := p.Database.Close()
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
