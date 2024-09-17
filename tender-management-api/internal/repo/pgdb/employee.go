package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"tender-management-api/internal/repo/repo_errors"
	"tender-management-api/pkg/postgres"

	"github.com/google/uuid"
)

type EmployeeRepo struct {
	*postgres.Postgres
}

func NewEmployeeRepo(pgdb *postgres.Postgres) *EmployeeRepo {
	return &EmployeeRepo{pgdb}
}

func (r *EmployeeRepo) GetEmployeeIdByUsername(ctx context.Context, username string) (string, error) {
	sqlReq, args, _ := r.SqlBuilder.
		Select("id").
		From("employee").
		Where("username = ?", username).
		ToSql()

	var employeeId string
	err := r.Database.QueryRow(sqlReq, args...).Scan(&employeeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repo_errors.ErrNotFound
		}

		return "", err
	}

	return employeeId, nil
}

func (r *EmployeeRepo) DoesOrganizationExistById(ctx context.Context, id string) (bool, error) {
	uuidForm, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	sqlReq, args, _ := r.SqlBuilder.
		Select("id").
		From("organization").
		Where("id = ?", uuidForm).
		ToSql()

	var organizationId string
	err = r.Database.QueryRow(sqlReq, args...).Scan(&organizationId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *EmployeeRepo) GetUserOrganizationIdByEmployeeId(ctx context.Context, employeeId string) (uuid.UUID, error) {
	uuidForm, err := uuid.Parse(employeeId)
	if err != nil {
		return uuid.Nil, err
	}

	sqlReq, args, _ := r.SqlBuilder.
		Select("organization_id").
		From("organization_responsible").
		Where("user_id = ?", uuidForm).
		ToSql()

	var organizationId uuid.UUID
	err = r.Database.QueryRow(sqlReq, args...).Scan(&organizationId)
	if err != nil {
		return uuid.Nil, err
	}

	return organizationId, nil
}

func (r *EmployeeRepo) IsEmployeeResponsible(ctx context.Context, employeeId string, organizationId uuid.UUID) (bool, error) {
	uuidForm, err := uuid.Parse(employeeId)
	if err != nil {
		return false, err
	}
	sqlReq, args, _ := r.SqlBuilder.
		Select("id").
		From("organization_responsible").
		Where("organization_id = ?", organizationId).
		Where("user_id = ?", uuidForm).
		ToSql()

	var id string
	err = r.Database.QueryRow(sqlReq, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *EmployeeRepo) DoesEmployeeExistsById(ctx context.Context, id string) (bool, error) {
	uuidForm, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	sqlReq, args, _ := r.SqlBuilder.
		Select("id").
		From("employee").
		Where("id = ?", uuidForm).
		ToSql()

	var uid uuid.UUID
	err = r.Database.QueryRow(sqlReq, args...).Scan(&uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
