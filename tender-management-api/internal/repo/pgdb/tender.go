package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"tender-management-api/internal/common"
	"tender-management-api/internal/entity"
	"tender-management-api/internal/repo/repo_errors"
	"tender-management-api/pkg/postgres"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type TenderRepo struct {
	*postgres.Postgres
}

func NewTenderRepo(pgdb *postgres.Postgres) *TenderRepo {
	return &TenderRepo{pgdb}
}

func (r *TenderRepo) CreateTender(ctx context.Context, input *entity.CreateTenderInput) (uuid.UUID, error) {
	tx, err := r.Database.Begin()
	if err != nil {
		return uuid.Nil, err
	}

	createTenderSql, args, _ := r.SqlBuilder.
		Insert("tender").
		Columns("status", "organization_id", "current_version").
		Values(common.Created, input.OrganizationId, 1).
		Suffix("RETURNING id").
		RunWith(tx).
		ToSql()

	var tenderId uuid.UUID
	err = tx.QueryRow(createTenderSql, args...).Scan(&tenderId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return uuid.Nil, err
		}

		return uuid.Nil, err
	}

	createVersionReq, args, _ := r.SqlBuilder.
		Insert("tender_version").
		Columns("name", "description", "service_type", "version", "tender_id").
		Values(input.Name, input.Description, input.ServiceType, 1, tenderId).
		RunWith(tx).
		ToSql()

	_, err = tx.Exec(createVersionReq, args...)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return uuid.Nil, err
		}

		return uuid.Nil, err
	}

	if err = tx.Commit(); err != nil {
		return uuid.Nil, err
	}

	return tenderId, nil
}

func (r *TenderRepo) GetTenderById(ctx context.Context, id string) (*entity.Tender, error) {
	getTenderSql, args, _ := r.SqlBuilder.
		Select("tender.created_at, tender.id, tender.status, tender.organization_id, tender_version.version, tender_version.name, tender_version.description, tender_version.service_type").
		From("tender").
		InnerJoin("tender_version on tender.id = tender_version.tender_id and tender.current_version = tender_version.version").
		Where("tender.id = ?", id).
		ToSql()

	var tender entity.Tender
	var createdAt time.Time
	row := r.Database.QueryRow(getTenderSql, args...)
	err := row.Scan(&createdAt, &tender.Id, &tender.Status, &tender.OrganizationId,
		&tender.Version, &tender.Name, &tender.Description, &tender.ServiceType)

	tender.CreatedAt = createdAt.Format(time.RFC3339)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &tender, repo_errors.ErrNotFound
		}

		return &tender, err
	}

	return &tender, nil
}

// Можно ругаться, если новая версия тендера не отличается от последней
// Но в задании такого требования нет + наверно не успею это сделать, поэтому оставлю как есть
func (r *TenderRepo) EditTenderById(ctx context.Context, id string, name string, description string, serviceType string) error {
	uuidForm, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	tx, err := r.Database.BeginTx(ctx, nil)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	updateVersionSql, args, _ := r.SqlBuilder.
		Update("tender").
		Set("current_version", squirrel.Expr("current_version + ?", 1)).
		Where("id = ?", id).
		Suffix("RETURNING current_version").
		RunWith(tx).
		ToSql()

	var currentVersion int
	err = tx.QueryRow(updateVersionSql, args...).Scan(&currentVersion)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	getOldValuesSql, args, _ := r.SqlBuilder.
		Select("name", "description", "service_type").
		From("tender_version").
		Where("tender_id = ?", id).
		Where("version = ?", currentVersion-1).
		RunWith(tx).
		ToSql()

	var prevName, prevDescr, prevServType string
	if err = tx.QueryRow(getOldValuesSql, args...).Scan(&prevName, &prevDescr, &prevServType); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	if name == "" {
		name = prevName
	}

	if description == "" {
		description = prevDescr
	}

	if serviceType == "" {
		serviceType = prevServType
	}

	createVersionSql, args, _ := r.SqlBuilder.
		Insert("tender_version").
		Columns("name", "description", "service_type", "version", "tender_id").
		Values(name, description, serviceType, currentVersion, uuidForm).
		RunWith(tx).
		ToSql()

	_, err = tx.Exec(createVersionSql, args...)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *TenderRepo) UpdateTenderStatusById(ctx context.Context, id string, newStatus string) error {
	uuidForm, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	updateStatusSql, args, _ := r.SqlBuilder.
		Update("tender").
		Set("status", newStatus).
		Where("id = ?", uuidForm).
		ToSql()

	_, err = r.Database.Exec(updateStatusSql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *TenderRepo) GetPublishedTenders(ctx context.Context, serviceTypes []string, pg *entity.PaginationInput) ([]entity.Tender, error) {
	builder := r.SqlBuilder.
		Select("tender.created_at, tender.id, tender.status, tender.organization_id, tender_version.version, tender_version.name, tender_version.description, tender_version.service_type").
		From("tender").
		InnerJoin("tender_version on tender.id = tender_version.tender_id and tender.current_version = tender_version.version").
		Where("status = ?", "Published")

	if len(serviceTypes) > 0 {
		builder = builder.Where(squirrel.Eq{"service_type": serviceTypes})
	}

	sqlReq, args, _ := builder.
		OrderBy("name ASC").
		Offset(uint64(pg.Offset)).
		Limit(uint64(pg.Limit)).
		ToSql()

	rows, err := r.Database.Query(sqlReq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenders := make([]entity.Tender, 0)
	for rows.Next() {
		var tender entity.Tender
		var createdAt time.Time
		if err := rows.Scan(&createdAt, &tender.Id, &tender.Status, &tender.OrganizationId,
			&tender.Version, &tender.Name, &tender.Description, &tender.ServiceType); err != nil {
			return tenders, err
		}
		tender.CreatedAt = createdAt.Format(time.RFC3339)
		tenders = append(tenders, tender)
	}
	if err = rows.Err(); err != nil {
		return tenders, err
	}

	return tenders, nil
}

func (r *TenderRepo) GetTendersByOrganizationId(ctx context.Context, organizationId uuid.UUID, pg *entity.PaginationInput) ([]entity.Tender, error) {
	sqlReq, args, _ := r.SqlBuilder.
		Select("tender.created_at, tender.id, tender.status, tender.organization_id, tender_version.version, tender_version.name, tender_version.description, tender_version.service_type").
		From("tender").
		InnerJoin("tender_version on tender.id = tender_version.tender_id and tender.current_version = tender_version.version").
		Where("organization_id = ?", organizationId.String()).
		OrderBy("name ASC").
		Offset(uint64(pg.Offset)).
		Limit(uint64(pg.Limit)).
		ToSql()

	rows, err := r.Database.Query(sqlReq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenders := make([]entity.Tender, 0)
	for rows.Next() {
		var tender entity.Tender
		var createdAt time.Time
		if err := rows.Scan(&createdAt, &tender.Id, &tender.Status, &tender.OrganizationId,
			&tender.Version, &tender.Name, &tender.Description, &tender.ServiceType); err != nil {
			return tenders, err
		}
		tender.CreatedAt = createdAt.Format(time.RFC3339)
		tenders = append(tenders, tender)
	}
	if err = rows.Err(); err != nil {
		return tenders, err
	}

	return tenders, nil
}

func (r *TenderRepo) RollbackTenderVersion(ctx context.Context, tenderId string, version int) error {
	uuidForm, err := uuid.Parse(tenderId)
	if err != nil {
		return err
	}

	tx, err := r.Database.BeginTx(ctx, nil)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	updateVersionInTenderTableSql, args, _ := r.SqlBuilder.
		Update("tender").
		Set("current_version", squirrel.Expr("current_version + ?", 1)).
		Where("id = ?", uuidForm).
		Suffix("RETURNING current_version").
		RunWith(tx).
		ToSql()

	var currentVersion int
	err = tx.QueryRow(updateVersionInTenderTableSql, args...).Scan(&currentVersion)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	updateVersionInVersionTableSql, args, _ := r.SqlBuilder.
		Update("tender_version").
		Set("version", currentVersion).
		Where("tender_id = ?", uuidForm).
		Where("version = ?", version).
		RunWith(tx).
		ToSql()

	if _, err = tx.Exec(updateVersionInVersionTableSql, args...); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
