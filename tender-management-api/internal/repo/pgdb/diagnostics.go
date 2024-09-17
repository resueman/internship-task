package pgdb

import (
	"tender-management-api/pkg/postgres"
)

type DiagnosticsRepo struct {
	*postgres.Postgres
}

func NewDiagnosticsRepo(pgdb *postgres.Postgres) *DiagnosticsRepo {
	return &DiagnosticsRepo{pgdb}
}

func (tr *DiagnosticsRepo) Ping() error {
	if err := tr.Database.Ping(); err != nil {
		return err
	}

	return nil
}
